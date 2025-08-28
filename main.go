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

// 企业微信消息结构
type WxMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
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

// 生成更人性化的消息内容
func generateMessage(quality string, eventTime string, aod string) string {
	var message string

	// 解析事件时间
	eventTimeFormatted := eventTime
	if eventTime != "" {
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", eventTime); err == nil {
			eventTimeFormatted = parsedTime.Format("2006年01月02日 15:04")
		}
	}

	switch quality {
	case "微微烧":
		message = fmt.Sprintf("🌅 火烧云预报 🌅\n当前时间：%s\n预测时间：%s\n\n虽然今晚的火烧云只是微微烧，但工作再忙也别忘记抬头看看天空哦～\n\n天空中的每一刻都是独特的，即使没有绚丽的火烧云，傍晚的天空也值得您驻足片刻，让眼睛和心灵都休息一下。\n\nAOD值：%s (空气质量还不错)",
			time.Now().Format("2006-01-02 15:04:05"), eventTimeFormatted, aod)
	case "小烧":
		message = fmt.Sprintf("🌅 火烧云预报 🌅\n当前时间：%s\n预测时间：%s\n\n今晚有小烧火烧云，工作累了就看看窗外吧！天空会给你一个小小的惊喜～\n\n记得给自己几分钟的休息时间，欣赏一下天边的温柔色彩。\n\nAOD值：%s",
			time.Now().Format("2006-01-02 15:04:05"), eventTimeFormatted, aod)
	case "小烧到中等烧":
		message = fmt.Sprintf("🌅 火烧云预报 🌅\n当前时间：%s\n预测时间：%s\n\n今晚的火烧云有小烧到中等烧，是时候暂时放下手头工作，看看天空的表演了！\n\n别让忙碌的生活错过了这些自然的小美好。\n\nAOD值：%s",
			time.Now().Format("2006-01-02 15:04:05"), eventTimeFormatted, aod)
	case "中等烧":
		message = fmt.Sprintf("🌅 火烧云预报 🌅\n当前时间：%s\n预测时间：%s\n\n今晚的火烧云有中等烧，值得您抽出片刻时间欣赏！\n\n工作是做不完的，但美丽的晚霞转瞬即逝，别错过这份天空的礼物。\n\nAOD值：%s",
			time.Now().Format("2006-01-02 15:04:05"), eventTimeFormatted, aod)
	case "中等烧到大烧":
		message = fmt.Sprintf("🌅 火烧云预报 🌅\n当前时间：%s\n预测时间：%s\n\n今晚的火烧云有中等烧到大烧，天空即将上演精彩表演！\n\n再忙也要记得抬头看看，让美丽的天空为您的日常增添一抹色彩。\n\nAOD值：%s",
			time.Now().Format("2006-01-02 15:04:05"), eventTimeFormatted, aod)
	case "大烧":
		message = fmt.Sprintf("🌅 火烧云预报 🌅\n当前时间：%s\n预测时间：%s\n\n今晚有大烧火烧云！别再埋头工作了，天空正在为您准备一场视觉盛宴！\n\n这是放松身心的完美时刻，别错过这份大自然的慷慨馈赠。\n\nAOD值：%s",
			time.Now().Format("2006-01-02 15:04:05"), eventTimeFormatted, aod)
	case "典型大烧":
		message = fmt.Sprintf("🌅 火烧云预报 🌅\n当前时间：%s\n预测时间：%s\n\n今晚将出现典型的大烧晚霞！工作可以等等，但这样的天空美景不可错过！\n\n给自己一个短暂的休息，让绚丽的天空为您充电。\n\nAOD值：%s",
			time.Now().Format("2006-01-02 15:04:05"), eventTimeFormatted, aod)
	case "优质大烧":
		message = fmt.Sprintf("🌅 火烧云预报 🌅\n当前时间：%s\n预测时间：%s\n\n今晚是优质大烧晚霞！这是拍照和放松的最佳时机！\n\n工作是重要的，但生活中的美好瞬间同样珍贵。别让这样的天空美景在忙碌中溜走。\n\nAOD值：%s",
			time.Now().Format("2006-01-02 15:04:05"), eventTimeFormatted, aod)
	case "世纪大烧":
		message = fmt.Sprintf("🌅 火烧云预报 🌅\n当前时间：%s\n预测时间：%s\n\n今晚的火烧云是世纪大烧！绝对是难得一见的壮观景象！\n\n无论多忙，请务必抽出时间欣赏这份天空的奇迹。工作可以等，但这样的美景可能一年只有几次！\n\nAOD值：%s",
			time.Now().Format("2006-01-02 15:04:05"), eventTimeFormatted, aod)
	default:
		message = fmt.Sprintf("🌅 火烧云预报 🌅\n当前时间：%s\n\n火烧云数据获取失败，但不管怎样，记得工作之余抬头看看天空，让眼睛和心灵都休息一下～",
			time.Now().Format("2006-01-02 15:04:05"))
	}
	return message
}

// 发送消息到企业微信 Webhook
func sendWxMsg(message string) error {
	// 直接在代码中写死 Webhook URL
	webhookURL := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=d87f2c6a-222a-43f6-91b8-5fbe251c8572"

	// 构造发送的消息
	wxMsg := WxMsg{
		MsgType: "text",
	}
	wxMsg.Text.Content = message

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

	// 生成消息内容
	message := generateMessage(quality, sunsetData.TbEventTime, sunsetData.TbAOD)

	// 发送消息到企业微信
	if err := sendWxMsg(message); err != nil {
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

// 定时任务：每天 5:30 PM 发送火烧云消息
func scheduleSunsetPush() {
	// 计算明天下午 5:30 的时间
	now := time.Now()
	nextRun := time.Date(now.Year(), now.Month(), now.Day(), 17, 30, 0, 0, time.Local)
	if now.After(nextRun) {
		// 如果当前时间已经过了 5:30，则推迟到明天 5:30
		nextRun = nextRun.Add(24 * time.Hour)
	}

	// 等待直到下一个定时推送
	duration := nextRun.Sub(now)
	log.Printf("下次推送将在 %s 后执行", duration)

	// 等待直到下一个定时推送
	time.Sleep(duration)

	// 执行定时任务
	for {
		// 获取火烧云数据
		sunsetData, err := getSunsetData()
		if err != nil {
			log.Printf("获取火烧云数据失败: %v", err)
			time.Sleep(1 * time.Hour) // 失败后等待1小时再试
			continue
		}

		// 提取质量数值
		qualityValue, err := extractQualityValue(sunsetData.TbQuality)
		if err != nil {
			log.Printf("解析质量值失败: %v", err)
			time.Sleep(1 * time.Hour) // 失败后等待1小时再试
			continue
		}

		// 判断火烧云等级
		quality := determineQualityLevel(qualityValue)

		// 生成消息内容
		message := generateMessage(quality, sunsetData.TbEventTime, sunsetData.TbAOD)

		// 发送消息到企业微信
		if err := sendWxMsg(message); err != nil {
			log.Printf("发送消息失败: %v", err)
			time.Sleep(1 * time.Hour) // 失败后等待1小时再试
			continue
		}

		log.Println("消息发送成功")

		// 等待24小时后再次执行
		time.Sleep(24 * time.Hour)
	}
}

func main() {
	// 启动定时任务
	go scheduleSunsetPush()

	// 启动 HTTP 服务，允许主动触发发送消息
	http.HandleFunc("/trigger-push", triggerPushHandler)
	log.Println("HTTP 服务已启动，监听端口 8080...")

	// 启动 HTTP 服务
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("启动 HTTP 服务失败: %v", err)
	}
}
