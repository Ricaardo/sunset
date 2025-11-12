package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// 配置结构
type Config struct {
	WebhookURL           string  // 企业微信 Webhook URL
	City                 string  // 城市
	Latitude             float64 // 纬度
	Longitude            float64 // 经度
	ScheduleHour         int     // 定时推送小时
	ScheduleMinute       int     // 定时推送分钟
	UseSunsetTime        bool    // 是否使用日落时间触发
	SunsetAdvanceMinutes int     // 日落前提前多少分钟推送（默认30分钟）
	Port                 string  // HTTP 服务端口
}

// 新的API响应数据结构
type SunsetData struct {
	TbAOD           string `json:"tb_aod"`            // 气溶胶光学厚度
	TbEventTime     string `json:"tb_event_time"`     // 事件时间
	TbQuality       string `json:"tb_quality"`        // 质量值，如 "0.047（微烧）"
	Status          string `json:"status"`            // API状态，如 "not_found"
	ImgSummary      string `json:"img_summary"`       // 图片摘要
	DisplayCityName string `json:"display_city_name"` // 显示的城市名称
}

// 企业微信消息结构 - 支持 markdown 格式
type WxMsg struct {
	MsgType  string `json:"msgtype"`
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

// 全局配置变量
var config Config

// 北京时区
var beijingLocation *time.Location

// 初始化配置
func initConfig() {
	var err error
	// 加载北京时区（东八区）
	beijingLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatalf("加载北京时区失败: %v", err)
	}

	// 从环境变量读取配置，如果没有则使用默认值
	config = Config{
		WebhookURL:           getEnv("WEBHOOK_URL", "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=58024cb2-a88f-4d70-84cd-7690ead0ead8"),
		City:                 getEnv("CITY", "上海市-上海"),
		Latitude:             getEnvFloat("LATITUDE", 31.2304),   // 上海纬度
		Longitude:            getEnvFloat("LONGITUDE", 121.4737), // 上海经度
		ScheduleHour:         getEnvInt("SCHEDULE_HOUR", 17),
		ScheduleMinute:       getEnvInt("SCHEDULE_MINUTE", 30),
		UseSunsetTime:        getEnvBool("USE_SUNSET_TIME", false),
		SunsetAdvanceMinutes: getEnvInt("SUNSET_ADVANCE_MINUTES", 30), // 默认提前30分钟
		Port:                 getEnv("PORT", "8080"),
	}

	log.Printf("配置加载完成:")
	log.Printf("  城市: %s", config.City)
	log.Printf("  坐标: %.4f, %.4f", config.Latitude, config.Longitude)
	log.Printf("  时区: 北京时间 (UTC+8)")
	if config.UseSunsetTime {
		log.Printf("  触发模式: 日落时间自动触发（提前 %d 分钟）", config.SunsetAdvanceMinutes)
	} else {
		log.Printf("  触发模式: 固定时间 %02d:%02d", config.ScheduleHour, config.ScheduleMinute)
	}
	log.Printf("  HTTP端口: %s", config.Port)
}

// 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 获取整数环境变量
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// 获取浮点数环境变量
func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// 获取布尔环境变量
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// 计算日落时间（使用改进的算法，参考NOAA方法）
func calculateSunsetTime(lat, lon float64, date time.Time) time.Time {
	// 将日期转换为儒略日
	jd := toJulianDay(date)

	// 计算从2000年1月1日12:00 UT起的儒略世纪数
	t := (jd - 2451545.0) / 36525.0

	// 计算太阳的平均黄经（度）
	l0 := math.Mod(280.46646+t*(36000.76983+t*0.0003032), 360.0)

	// 计算太阳的平均近点角（度）
	m := math.Mod(357.52911+t*(35999.05029-t*0.0001537), 360.0)
	mRad := m * math.Pi / 180.0

	// 计算地球轨道离心率
	e := 0.016708634 - t*(0.000042037+t*0.0000001267)

	// 计算太阳中心方程
	c := math.Sin(mRad)*(1.914602-t*(0.004817+t*0.000014)) +
		math.Sin(2*mRad)*(0.019993-t*0.000101) +
		math.Sin(3*mRad)*0.000289

	// 计算太阳的真黄经
	theta := l0 + c

	// 计算太阳的视黄经（考虑章动和光行差）
	omega := 125.04 - 1934.136*t
	lambda := theta - 0.00569 - 0.00478*math.Sin(omega*math.Pi/180.0)
	lambdaRad := lambda * math.Pi / 180.0

	// 计算黄赤交角（考虑章动）
	epsilon0 := 23.0 + 26.0/60.0 + 21.448/3600.0 - (46.8150*t+0.00059*t*t-0.001813*t*t*t)/3600.0
	epsilonRad := (epsilon0 + 0.00256*math.Cos(omega*math.Pi/180.0)) * math.Pi / 180.0

	// 计算太阳赤纬
	sinDec := math.Sin(epsilonRad) * math.Sin(lambdaRad)
	dec := math.Asin(sinDec)

	// 计算时间均衡差（分钟）
	y := math.Tan(epsilonRad/2.0) * math.Tan(epsilonRad/2.0)
	eqTime := 4.0 * (y*math.Sin(2*l0*math.Pi/180.0) -
		2*e*math.Sin(mRad) +
		4*e*y*math.Sin(mRad)*math.Cos(2*l0*math.Pi/180.0) -
		0.5*y*y*math.Sin(4*l0*math.Pi/180.0) -
		1.25*e*e*math.Sin(2*mRad)) * 180.0 / math.Pi

	// 计算日落时角（考虑大气折射，使用-0.833度）
	latRad := lat * math.Pi / 180.0
	zenith := 90.833 * math.Pi / 180.0 // 90度50分（0.833度）

	cosHA := (math.Cos(zenith) - math.Sin(latRad)*math.Sin(dec)) / (math.Cos(latRad) * math.Cos(dec))

	// 检查是否有日出日落（极昼极夜情况）
	if cosHA > 1.0 {
		cosHA = 1.0 // 极夜
	} else if cosHA < -1.0 {
		cosHA = -1.0 // 极昼
	}

	// 计算时角（度）
	ha := math.Acos(cosHA) * 180.0 / math.Pi

	// 计算日落时间（UTC，分钟）
	// 日落时间 = 正午(720分钟) + 时角偏移 - 经度修正 - 时间均差
	sunsetMinutes := 720.0 + 4.0*ha - 4.0*lon - eqTime

	// 转换为UTC小时
	utcHours := sunsetMinutes / 60.0

	// 转换为北京时间（UTC+8）
	beijingHours := utcHours + 8.0

	// 处理跨天情况
	for beijingHours < 0 {
		beijingHours += 24
		date = date.Add(-24 * time.Hour)
	}
	for beijingHours >= 24 {
		beijingHours -= 24
		date = date.Add(24 * time.Hour)
	}

	hour := int(beijingHours)
	minute := int((beijingHours - float64(hour)) * 60)
	second := int(((beijingHours-float64(hour))*60 - float64(minute)) * 60)

	return time.Date(date.Year(), date.Month(), date.Day(), hour, minute, second, 0, beijingLocation)
}

// 转换为儒略日
func toJulianDay(t time.Time) float64 {
	y := t.Year()
	m := int(t.Month())
	d := t.Day()

	if m <= 2 {
		y--
		m += 12
	}

	a := y / 100
	b := 2 - a + a/4

	jd := float64(int(365.25*float64(y+4716))) + float64(int(30.6001*float64(m+1))) + float64(d) + float64(b) - 1524.5

	return jd
}

// 获取火烧云数据
func getSunsetData() (SunsetData, error) {
	// 尝试使用详细API接口
	data, err := fetchFromDetailedAPI()
	if err == nil {
		return data, nil
	}

	log.Printf("详细API失败: %v, 尝试简洁API接口", err)

	// 如果详细API失败，尝试简洁API
	data, err = fetchFromSimpleAPI()
	if err != nil {
		return SunsetData{}, fmt.Errorf("两个API接口都失败了 - 详细API和简洁API")
	}

	return data, nil
}

// 从详细API获取数据
func fetchFromDetailedAPI() (SunsetData, error) {
	// 使用配置的城市构建 URL
	// 城市格式保持 "上海市-上海"（带连字符），直接进行 URL 编码
	// 参考正确的API请求：query_city=%E4%B8%8A%E6%B5%B7%E5%B8%82-%E4%B8%8A%E6%B5%B7

	// 正确进行 URL 编码，使用 EC 模型（更准确）
	apiURL := fmt.Sprintf("https://sunsetbot.top/detailed/?query_id=8454963&intend=select_city&query_city=%s&model=EC&event_date=None&event=set_1&times=None",
		url.QueryEscape(config.City))

	log.Printf("请求详细API: %s", apiURL)
	log.Printf("城市参数: %s", config.City)

	// 创建带超时的 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return SunsetData{}, fmt.Errorf("创建请求失败: %w", err)
	}

	// 添加必要的请求头，模拟浏览器请求
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://sunsetbot.top/detailed/")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	// 发起 HTTP 请求
	resp, err := client.Do(req)
	if err != nil {
		return SunsetData{}, fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return SunsetData{}, fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	// 解析返回的数据
	var data SunsetData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return SunsetData{}, fmt.Errorf("JSON 解析失败: %w", err)
	}

	// 记录接收到的数据（用于调试）
	log.Printf("接收到的数据: status='%s', display_city_name='%s', tb_quality='%s', tb_event_time='%s', tb_aod='%s'",
		data.Status, data.DisplayCityName, data.TbQuality, data.TbEventTime, data.TbAOD)

	// 检查API状态
	if data.Status == "not_found" {
		return SunsetData{}, fmt.Errorf("API 未找到该城市的数据: %s (城市名: %s)。请检查城市名称格式是否正确，或该城市可能暂无预报数据",
			data.ImgSummary, config.City)
	}

	// 验证必要字段
	if data.TbQuality == "" {
		return SunsetData{}, fmt.Errorf("API 返回的 tb_quality 字段为空，status='%s', 可能该城市暂无数据或API格式发生变化", data.Status)
	}

	return data, nil
}

// 从简洁API获取数据（备用方案）
func fetchFromSimpleAPI() (SunsetData, error) {
	// 简洁API接口：通过Cookie传递城市信息
	// 参考: https://sunsetbot.top/?query_id=1344491&intend=select_city&query_city=&event_date=None&event=set_1&times=None&model=EC

	apiURL := "https://sunsetbot.top/?query_id=1344491&intend=select_city&query_city=&event_date=None&event=set_1&times=None&model=EC"

	log.Printf("请求简洁API: %s", apiURL)

	// 创建带超时的 HTTP 客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return SunsetData{}, fmt.Errorf("创建请求失败: %w", err)
	}

	// 提取城市名（去掉省市后缀）- 如 "上海市-上海" -> "上海"
	cityParts := strings.Split(config.City, "-")
	cityName := config.City
	if len(cityParts) > 1 {
		cityName = cityParts[1]
	}

	// 添加必要的请求头和Cookie
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://sunsetbot.top/")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	// 添加城市Cookie（URL编码的城市名）
	cookieValue := url.QueryEscape(fmt.Sprintf("\"%s\"", cityName))
	req.Header.Set("Cookie", fmt.Sprintf("city_name=%s", cookieValue))

	log.Printf("使用城市Cookie: city_name=\"%s\"", cityName)

	// 发起 HTTP 请求
	resp, err := client.Do(req)
	if err != nil {
		return SunsetData{}, fmt.Errorf("HTTP 请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return SunsetData{}, fmt.Errorf("API 返回错误状态码: %d", resp.StatusCode)
	}

	// 解析返回的数据
	var data SunsetData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return SunsetData{}, fmt.Errorf("JSON 解析失败: %w", err)
	}

	// 记录接收到的数据（用于调试）
	log.Printf("简洁API返回数据: status='%s', display_city_name='%s', tb_quality='%s'",
		data.Status, data.DisplayCityName, data.TbQuality)

	// 检查API状态
	if data.Status != "ok" {
		return SunsetData{}, fmt.Errorf("简洁API返回状态异常: status='%s'", data.Status)
	}

	// 验证必要字段
	if data.TbQuality == "" {
		return SunsetData{}, fmt.Errorf("简洁API返回的 tb_quality 字段为空")
	}

	return data, nil
}

// 从质量字符串中提取数值部分
func extractQualityValue(qualityStr string) (float64, error) {
	// 检查字符串是否为空
	if qualityStr == "" {
		return 0, fmt.Errorf("质量字符串为空")
	}

	// 提取数字部分（去掉括号和中文描述）
	parts := strings.Split(qualityStr, "（")
	if len(parts) == 0 || parts[0] == "" {
		return 0, fmt.Errorf("无效的质量字符串格式: %s", qualityStr)
	}

	// 清理字符串（去除空格）
	numStr := strings.TrimSpace(parts[0])
	if numStr == "" {
		return 0, fmt.Errorf("数值部分为空: %s", qualityStr)
	}

	// 转换字符串为浮点数
	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("无法解析数值 '%s': %w", numStr, err)
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
	// 获取当前北京时间
	now := time.Now().In(beijingLocation)
	currentTime := now.Format("2006-01-02 15:04:05")

	// 解析事件时间（日落时间），如果为空则使用当前日期的日落时间
	var eventTimeFormatted string
	if eventTime != "" {
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", eventTime); err == nil {
			eventTimeFormatted = parsedTime.Format("15:04")
		} else {
			// 如果解析失败，使用计算的日落时间
			sunsetTime := calculateSunsetTime(config.Latitude, config.Longitude, now)
			eventTimeFormatted = sunsetTime.Format("15:04")
		}
	} else {
		// 如果没有事件时间，使用计算的日落时间
		sunsetTime := calculateSunsetTime(config.Latitude, config.Longitude, now)
		eventTimeFormatted = sunsetTime.Format("15:04")
	}

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

	// 创建Markdown格式的富文本消息 - 温馨版
	message := fmt.Sprintf(`## %s %s

> <font color="%s">**质量等级：%s**</font>
> %s

---

### 📅 今日预报

**日落时间**：今天 %s
**空气质量**：%s
**推送时间**：%s

---

%s

---

%s

<font color="comment">💬 来自天空的问候 · 数据来源 SunsetBot</font>`,
		title, emoji,
		color, quality, qualityBar,
		eventTimeFormatted,
		aod,
		currentTime,
		getQualityDescription(quality),
		getRandomTip())

	return message
}

// 根据质量等级获取描述文本
func getQualityDescription(quality string) string {
	switch quality {
	case "微微烧":
		return "> 🌤️ 虽然今晚的火烧云只是微微烧，但天空的每一刻都是独特的。\n> \n> 工作再忙，也别忘记抬头看看天空，让眼睛和心灵都休息一下。即使没有绚丽的火烧云，傍晚的天空也值得您驻足片刻。"
	case "小烧":
		return "> 🌇 今晚有小烧火烧云，天边会有淡淡的色彩。\n> \n> 工作累了就看看窗外吧，天空会给你一个小小的惊喜。记得给自己几分钟的休息时间，欣赏一下天边的温柔色彩。"
	case "小烧到中等烧":
		return "> ⛅ 今晚的火烧云有小烧到中等烧，天空将呈现温柔的渐变色彩。\n> \n> 是时候暂时放下手头工作，看看天空的表演了。别让忙碌的生活错过了这些自然的小美好。"
	case "中等烧":
		return "> 🔥 今晚的火烧云有中等烧，色彩层次丰富，值得您抽出片刻时间欣赏！\n> \n> 工作是做不完的，但美丽的晚霞转瞬即逝。别错过这份天空的礼物，让美景为您的日常增添一抹温暖的色彩。"
	case "中等烧到大烧":
		return "> 🌅 今晚的火烧云有中等烧到大烧，天空即将上演精彩表演！\n> \n> 再忙也要记得抬头看看，这样的美景能让心情瞬间变好。建议找个视野好的地方，好好享受这份来自天空的馈赠。"
	case "大烧":
		return "> 🌄 今晚有大烧火烧云！天空正在为您准备一场视觉盛宴！\n> \n> 别再埋头工作了，这是放松身心的完美时刻。叫上三五好友，一起分享这份大自然的慷慨馈赠，让美景治愈疲惫的心灵。"
	case "典型大烧":
		return "> ✨ 今晚将出现典型的大烧晚霞！色彩饱满、层次丰富！\n> \n> 工作可以等等，但这样的天空美景不可错过。强烈建议找个高处，静静欣赏这场视觉盛宴。给自己一个短暂的休息，让绚丽的天空为您充电。"
	case "优质大烧":
		return "> 📸 今晚是优质大烧晚霞！这是拍照和放松的最佳时机！\n> \n> 天空将呈现极其绚丽的色彩，记得带上相机或手机。工作是重要的，但生活中的美好瞬间同样珍贵。别让这样的天空美景在忙碌中溜走！"
	case "世纪大烧":
		return "> 🌌 **今晚的火烧云是世纪大烧！这是极其罕见的壮观景象！**\n> \n> 一年可能只有几次这样的机会，无论多忙，请务必抽出时间欣赏这份天空的奇迹。\n> \n> 💡 **温馨建议**：\n> • 提前15分钟找个视野开阔的高处\n> • 准备好手机或相机，记录这珍贵时刻\n> • 叫上朋友或家人一起，分享这份感动\n> • 放下手机专注欣赏几分钟，用眼睛和心感受\n> \n> **错过今天，下次不知道要等多久了！**"
	default:
		return "> 🌅 虽然火烧云数据获取失败，但傍晚的天空依然值得一看。\n> \n> 不管怎样，记得工作之余抬头看看天空，让眼睛和心灵都休息一下。有时候，最美的风景就在不经意间。"
	}
}

// 获取随机小贴士
func getRandomTip() string {
	tips := []string{
		"💡 **小贴士**：手机拍晚霞时，点击屏幕最亮处降低曝光，色彩会更饱满。打开HDR模式效果更佳哦～",
		"💡 **小贴士**：观赏晚霞的最佳位置是视野开阔的高处或水边，能看到完整的天空画卷。提前踩点更不会错过美景～",
		"💡 **小贴士**：火烧云通常在日落前后20-30分钟最美，建议提前10分钟到位。美景转瞬即逝，别迟到啦！",
		"💡 **小贴士**：傍晚远眺天空可以有效缓解眼疲劳哦。每天给自己5分钟「天空时间」，眼睛和心灵都会感谢你～",
		"💡 **小贴士**：空气质量好 + 云层适中（30-70%）= 绚丽晚霞。雨后放晴是观赏良机，记得把握！",
		"💡 **小贴士**：叫上朋友或家人一起看晚霞吧！分享美景能让快乐加倍。发个朋友圈，当大家的「天空播报员」～",
		"💡 **小贴士**：拍摄时加入建筑、树木或人物作为前景，画面会更有层次感。试试黄金分割线构图，效果惊艳！",
	}

	// 使用当前日期作为种子，确保每天显示相同的小贴士
	now := time.Now()
	index := now.YearDay() % len(tips)
	return tips[index]
}

// 发送富文本卡片消息到企业微信 Webhook
func sendWxMarkdownMsg(message string) error {
	// 使用配置的 Webhook URL
	webhookURL := config.WebhookURL

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

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("企业微信返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// 处理主动触发推送的HTTP请求
func triggerPushHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("收到主动触发推送请求")

	// 执行推送任务
	if err := executePushTask(); err != nil {
		http.Error(w, fmt.Sprintf("推送失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "success",
		"message":   "消息发送成功",
		"timestamp": time.Now().In(beijingLocation).Format("2006-01-02 15:04:05"),
	})
}

// 处理健康检查请求
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().In(beijingLocation).Format("2006-01-02 15:04:05"),
		"timezone":  "Asia/Shanghai (UTC+8)",
	})
}

// 处理配置查询请求
func configHandler(w http.ResponseWriter, r *http.Request) {
	nextPushTime := getNextPushTime()

	response := map[string]interface{}{
		"city":            config.City,
		"latitude":        config.Latitude,
		"longitude":       config.Longitude,
		"schedule_hour":   config.ScheduleHour,
		"schedule_minute": config.ScheduleMinute,
		"use_sunset_time": config.UseSunsetTime,
		"timezone":        "Asia/Shanghai (UTC+8)",
		"next_push_time":  nextPushTime.Format("2006-01-02 15:04:05"),
		"current_time":    time.Now().In(beijingLocation).Format("2006-01-02 15:04:05"),
	}

	// 如果使用日落时间模式，返回额外信息
	if config.UseSunsetTime {
		now := time.Now().In(beijingLocation)
		sunsetTime := calculateSunsetTime(config.Latitude, config.Longitude, now)
		if now.After(sunsetTime) {
			sunsetTime = calculateSunsetTime(config.Latitude, config.Longitude, now.Add(24*time.Hour))
		}
		response["sunset_advance_minutes"] = config.SunsetAdvanceMinutes
		response["next_sunset_time"] = sunsetTime.Format("2006-01-02 15:04:05")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// 处理日落时间查询请求
func sunsetTimeHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now().In(beijingLocation)
	todaySunset := calculateSunsetTime(config.Latitude, config.Longitude, now)

	// 如果今天的日落时间已过，显示明天的
	if now.After(todaySunset) {
		todaySunset = calculateSunsetTime(config.Latitude, config.Longitude, now.Add(24*time.Hour))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sunset_time":  todaySunset.Format("2006-01-02 15:04:05"),
		"current_time": now.Format("2006-01-02 15:04:05"),
		"latitude":     config.Latitude,
		"longitude":    config.Longitude,
		"city":         config.City,
	})
}

// 执行推送任务的核心逻辑
func executePushTask() error {
	// 获取火烧云数据
	sunsetData, err := getSunsetData()
	if err != nil {
		return fmt.Errorf("获取火烧云数据失败: %v", err)
	}

	// 提取质量数值
	qualityValue, err := extractQualityValue(sunsetData.TbQuality)
	if err != nil {
		return fmt.Errorf("解析质量值失败: %v", err)
	}

	// 判断火烧云等级
	quality := determineQualityLevel(qualityValue)

	// 生成富文本消息内容
	message := generateMarkdownMessage(quality, sunsetData.TbEventTime, sunsetData.TbAOD)

	// 发送消息到企业微信
	if err := sendWxMarkdownMsg(message); err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}

	log.Printf("消息发送成功 - 质量等级: %s", quality)
	return nil
}

// 计算下次推送时间
func getNextPushTime() time.Time {
	now := time.Now().In(beijingLocation)

	if config.UseSunsetTime {
		// 使用日落时间触发
		sunsetTime := calculateSunsetTime(config.Latitude, config.Longitude, now)

		// 提前指定分钟数推送
		pushTime := sunsetTime.Add(-time.Duration(config.SunsetAdvanceMinutes) * time.Minute)

		// 如果今天的推送时间已过，计算明天的日落时间
		if now.After(pushTime) {
			tomorrow := now.Add(24 * time.Hour)
			sunsetTime = calculateSunsetTime(config.Latitude, config.Longitude, tomorrow)
			pushTime = sunsetTime.Add(-time.Duration(config.SunsetAdvanceMinutes) * time.Minute)
		}

		log.Printf("使用日落时间触发模式 - 日落时间: %s, 推送时间: %s (提前 %d 分钟)",
			sunsetTime.Format("2006-01-02 15:04:05"),
			pushTime.Format("2006-01-02 15:04:05"),
			config.SunsetAdvanceMinutes)
		return pushTime
	} else {
		// 使用固定时间触发
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), config.ScheduleHour, config.ScheduleMinute, 0, 0, beijingLocation)

		// 如果当前时间已经过了设定时间，则推迟到明天
		if now.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}

		log.Printf("使用固定时间触发模式 - 下次推送时间: %s", nextRun.Format("2006-01-02 15:04:05"))
		return nextRun
	}
}

// 定时任务：发送火烧云消息
func scheduleSunsetPush() {
	for {
		// 计算下次推送时间
		nextRun := getNextPushTime()
		now := time.Now().In(beijingLocation)

		// 计算等待时间
		duration := nextRun.Sub(now)
		log.Printf("距离下次推送还有: %s (将在 %s 执行)", duration, nextRun.Format("2006-01-02 15:04:05"))

		// 等待直到下一个定时推送
		time.Sleep(duration)

		// 执行推送任务
		log.Println("开始执行定时推送任务...")
		if err := executePushTask(); err != nil {
			log.Printf("推送任务失败: %v", err)
			// 失败后等待10分钟再计算下次推送时间
			time.Sleep(10 * time.Minute)
		} else {
			// 成功后等待1分钟，防止重复推送
			time.Sleep(1 * time.Minute)
		}
	}
}

func main() {
	// 初始化配置
	initConfig()

	log.Println("========================================")
	log.Println("火烧云推送服务启动中...")
	log.Println("========================================")

	// 启动定时任务
	go scheduleSunsetPush()

	// 注册 HTTP 路由
	http.HandleFunc("/trigger-push", triggerPushHandler) // 主动触发推送
	http.HandleFunc("/health", healthCheckHandler)       // 健康检查
	http.HandleFunc("/config", configHandler)            // 查询配置
	http.HandleFunc("/sunset-time", sunsetTimeHandler)   // 查询日落时间

	// 启动 HTTP 服务
	serverAddr := ":" + config.Port
	log.Printf("HTTP 服务已启动，监听端口 %s", config.Port)
	log.Println("可用的 API 端点:")
	log.Printf("  - GET/POST  http://localhost:%s/trigger-push   主动触发推送", config.Port)
	log.Printf("  - GET       http://localhost:%s/health         健康检查", config.Port)
	log.Printf("  - GET       http://localhost:%s/config         查询配置", config.Port)
	log.Printf("  - GET       http://localhost:%s/sunset-time    查询日落时间", config.Port)
	log.Println("========================================")

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("启动 HTTP 服务失败: %v", err)
	}
}
