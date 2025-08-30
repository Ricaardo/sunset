package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 新的API响应数据结构
type SunsetData struct {
	TbAOD       string `json:"tb_aod"`        // 气溶胶光学厚度
	TbEventTime string `json:"tb_event_time"` // 事件时间
	TbQuality   string `json:"tb_quality"`    // 质量值，如 "0.047（微烧）"
}

// 企业微信消息结构 - 支持 markdown 格式
type WxMsg struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

// 获取火烧云数据
func getSunsetData() (SunsetData, error) {
	url := "https://sunsetbot.top/detailed/?query_id=4624758&intend=select_city&query_city=%E4%B8%8A%E6%B5%B7%E5%B8%82-%E4%B8%8A%E6%B5%B7&model=GFS&event_date=None&event=set_1&times=None"

	// 发起 HTTP 请求
	resp, err := http.Get(url)
	if err != nil {
		return SunsetData{}, err
	}
	defer resp.Body.Close()

	// 解析返回的数据
	var data SunsetData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return SunsetData{}, err
	}

	return data, nil
}

// 从质量字符串中提取数值部分
func extractQualityValue(qualityStr string) (float64, error) {
	// 提取数字部分（去掉括号和中文描述）
	parts := strings.Split(qualityStr, "（")
	if len(parts) == 0 {
		return 0, fmt.Errorf("无效的质量字符串格式")
	}

	// 转换字符串为浮点数
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}

// 根据质量值判断火烧云等级
func determineQualityLevel(qualityValue float64) string {
	switch {
	case qualityValue < 0.05:
		return "微微烧"
	case qualityValue < 0.1:
		return "小烧"
	case qualityValue < 0.2:
		return "小烧到中等烧"
	case qualityValue < 0.3:
		return "中等烧"
	case qualityValue < 0.4:
		return "中等烧到大烧"
	case qualityValue < 0.5:
		return "大烧"
	case qualityValue < 0.6:
		return "典型大烧"
	case qualityValue < 0.7:
		return "优质大烧"
	default:
		return "世纪大烧"
	}
}

// 生成富文本卡片消息内容 (Markdown格式)
func generateMarkdownMessage(quality string, eventTime string, aod string) string {
	// 解析事件时间
	eventTimeFormatted := eventTime
	if eventTime != "" {
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", eventTime); err == nil {
			eventTimeFormatted = parsedTime.Format("2006年01月02日 15:04")
		}
	}

	// 获取当前时间
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// 根据质量等级选择不同的表情和颜色
	var emoji, color, title, qualityBar string
	switch quality {
	case "微微烧":
		emoji = "🌤️"
		color = "#808080" // 灰色
		title = "今日晚霞预报"
		qualityBar = "▁▂▃▅▆▇"
	case "小烧":
		emoji = "🌇"
		color = "#1E90FF" // 蓝色
		title = "今日晚霞预报"
		qualityBar = "▁▂▃▅▆▇"
	case "小烧到中等烧":
		emoji = "⛅"
		color = "#32CD32" // 绿色
		title = "今日晚霞预报"
		qualityBar = "▁▂▃▅▆▇"
	case "中等烧":
		emoji = "🔥"
		color = "#FFA500" // 橙色
		title = "今日晚霞预报"
		qualityBar = "▁▂▃▅▆▇"
	case "中等烧到大烧":
		emoji = "🌅"
		color = "#FF4500" // 橙红色
		title = "今日晚霞预报"
		qualityBar = "▁▂▃▅▆▇"
	case "大烧":
		emoji = "🌄"
		color = "#FF0000" // 红色
		title = "今日晚霞预报"
		qualityBar = "▁▂▃▅▆▇"
	case "典型大烧":
		emoji = "✨"
		color = "#9370DB" // 紫色
		title = "今日晚霞预报"
		qualityBar = "▁▂▃▅▆▇"
	case "优质大烧":
		emoji = "📸"
		color = "#EE82EE" // 紫罗兰色
		title = "今日晚霞预报"
		qualityBar = "▁▂▃▅▆▇"
	case "世纪大烧":
		emoji = "🌌"
		color = "#4B0082" // 靛蓝色
		title = "今日晚霞预报"
		qualityBar = "▁▂▃▅▆▇"
	default:
		emoji = "🌅"
		color = "#000000" // 黑色
		title = "今日晚霞预报"
		qualityBar = "▁▂▃▅▆▇"
	}

	// 创建Markdown格式的富文本消息
	message := fmt.Sprintf(`<font color="%s">**%s %s**</font>

<font color="%s">**预报质量:** %s %s</font>

**预测时间:** %s
**当前时间:** %s
**空气质量:** %s

%s

%s`,
		color, title, emoji,
		color, quality, qualityBar,
		eventTimeFormatted,
		currentTime,
		aod,
		getQualityDescription(quality),
		getRandomTip())

	return message
}

// 根据质量等级获取描述文本
func getQualityDescription(quality string) string {
	switch quality {
	case "微微烧":
		return "> 🌤️ 虽然今晚的火烧云只是微微烧，但工作再忙也别忘记抬头看看天空哦～\n> \n> 天空中的每一刻都是独特的，即使没有绚丽的火烧云，傍晚的天空也值得您驻足片刻，让眼睛和心灵都休息一下。"
	case "小烧":
		return "> 🌇 今晚有小烧火烧云，工作累了就看看窗外吧！天空会给你一个小小的惊喜～\n> \n> 记得给自己几分钟的休息时间，欣赏一下天边的温柔色彩。"
	case "小烧到中等烧":
		return "> ⛅ 今晚的火烧云有小烧到中等烧，是时候暂时放下手头工作，看看天空的表演了！\n> \n> 别让忙碌的生活错过了这些自然的小美好。"
	case "中等烧":
		return "> 🔥 今晚的火烧云有中等烧，值得您抽出片刻时间欣赏！\n> \n> 工作是做不完的，但美丽的晚霞转瞬即逝，别错过这份天空的礼物。"
	case "中等烧到大烧":
		return "> 🌅 今晚的火烧云有中等烧到大烧，天空即将上演精彩表演！\n> \n> 再忙也要记得抬头看看，让美丽的天空为您的日常增添一抹色彩。"
	case "大烧":
		return "> 🌄 今晚有大烧火烧云！别再埋头工作了，天空正在为您准备一场视觉盛宴！\n> \n> 这是放松身心的完美时刻，别错过这份大自然的慷慨馈赠。"
	case "典型大烧":
		return "> ✨ 今晚将出现典型的大烧晚霞！工作可以等等，但这样的天空美景不可错过！\n> \n> 给自己一个短暂的休息，让绚丽的天空为您充电。"
	case "优质大烧":
		return "> 📸 今晚是优质大烧晚霞！这是拍照和放松的最佳时机！\n> \n> 工作是重要的，但生活中的美好瞬间同样珍贵。别让这样的天空美景在忙碌中溜走。"
	case "世纪大烧":
		return "> 🌌 今晚的火烧云是世纪大烧！绝对是难得一见的壮观景象！\n> \n> 无论多忙，请务必抽出时间欣赏这份天空的奇迹。工作可以等，但这样的美景可能一年只有几次！"
	default:
		return "> 🌅 火烧云数据获取失败，但不管怎样，记得工作之余抬头看看天空，让眼睛和心灵都休息一下～"
	}
}

// 获取随机小贴士
func getRandomTip() string {
	tips := []string{
		"💡 **小贴士**: 观赏晚霞的最佳位置是视野开阔的高处或水边",
		"💡 **小贴士**: 使用手机的专业模式拍摄晚霞，调整白平衡和曝光可以获得更好效果",
		"💡 **小贴士**: 傍晚时分是放松眼睛的好时机，远眺天空可以缓解眼部疲劳",
		"💡 **小贴士**: 火烧云通常出现在日落前后20-30分钟，别错过最佳观赏时间",
		"💡 **小贴士**: 空气质量好的日子，晚霞通常更加绚丽",
		"💡 **小贴士**: 找个伴一起欣赏晚霞，分享美景会让心情更加愉悦",
		"💡 **小贴士**: 工作间隙看看天空，可以帮助放松大脑，提高工作效率",
	}

	// 使用当前日期作为种子，确保每天显示相同的小贴士
	now := time.Now()
	index := now.YearDay() % len(tips)
	return tips[index]
}

// 发送富文本卡片消息到企业微信 Webhook
func sendWxMarkdownMsg(message string) error {
	// 直接在代码中写死 Webhook URL
	webhookURL := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=58024cb2-a88f-4d70-84cd-7690ead0ead8"

	// 构造发送的消息
	wxMsg := WxMsg{
		MsgType: "markdown",
	}
	wxMsg.Markdown.Content = message

	// 发送消息
	msgBody, _ := json.Marshal(wxMsg)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(msgBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// 处理主动触发推送的HTTP请求
func triggerPushHandler(w http.ResponseWriter, r *http.Request) {
	// 获取火烧云数据
	sunsetData, err := getSunsetData()
	if err != nil {
		http.Error(w, fmt.Sprintf("获取火烧云数据失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 提取质量数值
	qualityValue, err := extractQualityValue(sunsetData.TbQuality)
	if err != nil {
		http.Error(w, fmt.Sprintf("解析质量值失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 判断火烧云等级
	quality := determineQualityLevel(qualityValue)

	// 生成富文本消息内容
	message := generateMarkdownMessage(quality, sunsetData.TbEventTime, sunsetData.TbAOD)

	// 发送消息到企业微信
	if err := sendWxMarkdownMsg(message); err != nil {
		http.Error(w, fmt.Sprintf("发送消息失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "消息发送成功",
		"quality": quality,
	})
}

// 推送一次火烧云消息
func pushSunsetMsg() error {
	sunsetData, err := getSunsetData()
	if err != nil {
		return fmt.Errorf("获取火烧云数据失败: %v", err)
	}
	qualityValue, err := extractQualityValue(sunsetData.TbQuality)
	if err != nil {
		return fmt.Errorf("解析质量值失败: %v", err)
	}
	quality := determineQualityLevel(qualityValue)
	message := generateMarkdownMessage(quality, sunsetData.TbEventTime, sunsetData.TbAOD)
	if err := sendWxMarkdownMsg(message); err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}
	return nil
}

// 优化后的定时任务：每天指定时间推送火烧云消息
func scheduleSunsetPush(hour, min int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("定时任务 panic: %v", r)
		}
	}()
	for {
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, time.Local)
		if now.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}
		duration := nextRun.Sub(now)
		log.Printf("下次推送将在 %s (%s) 后执行", nextRun.Format("2006-01-02 15:04:05"), duration)
		time.Sleep(duration)
		if err := pushSunsetMsg(); err != nil {
			log.Printf("定时推送失败: %v", err)
			time.Sleep(1 * time.Hour)
			continue
		}
		log.Println("定时推送成功")
		time.Sleep(24 * time.Hour)
	}
}

func main() {
	go scheduleSunsetPush(17, 30) // 可改为 go scheduleSunsetPush(时, 分) 方便测试
	http.HandleFunc("/trigger-push", triggerPushHandler)
	log.Println("HTTP 服务已启动，监听端口 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("启动 HTTP 服务失败: %v", err)
	}
}
