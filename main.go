package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// OpenMeteo 的天气数据结构
type WeatherData struct {
	Hourly struct {
		Temperature_2m []float64 `json:"temperature_2m"`
		Cloudcover     []float64 `json:"cloudcover"`
	} `json:"hourly"`
	Current struct {
		Sunset time.Time `json:"sunset"`
	} `json:"current"`
}

// 企业微信消息结构
type WxMsg struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

// 获取上海的天气数据
func getWeather() (WeatherData, error) {
	// Open-Meteo API 的请求 URL
	url := "https://api.open-meteo.com/v1/forecast?latitude=31.2304&longitude=121.4737&hourly=temperature_2m,cloudcover"

	// 发起 HTTP 请求
	resp, err := http.Get(url)
	if err != nil {
		return WeatherData{}, err
	}
	defer resp.Body.Close()

	// 解析返回的天气数据
	var data WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return WeatherData{}, err
	}

	return data, nil
}

// 判断是否有晚霞并返回晚霞等级
func getSunsetQuality(cloudCoverage float64, sunsetTime time.Time) string {
	// 判断当前时间是否接近日落时间
	now := time.Now()
	if now.After(sunsetTime) && now.Before(sunsetTime.Add(time.Hour)) {
		// 如果是接近日落时间
		if cloudCoverage > 50 {
			return "无晚霞"
		} else if cloudCoverage > 30 {
			return "小烧"
		} else if cloudCoverage > 10 {
			return "中烧"
		}
		return "大烧"
	}
	return "无晚霞"
}

// 生成人性化提醒的消息内容
func generateMessage(quality string) string {
	var message string
	switch quality {
	case "无晚霞":
		message = fmt.Sprintf("当前时间：%s\n今天的天气不太适合观赏晚霞，云层比较多，建议找点别的活动吧！🌥", time.Now().Format("2006-01-02 15:04:05"))
	case "小烧":
		message = fmt.Sprintf("当前时间：%s\n今晚的晚霞小有点烧，虽然不算壮丽，但也值得期待！如果有空，记得去看看天边的美丽色彩哦！🌇", time.Now().Format("2006-01-02 15:04:05"))
	case "中烧":
		message = fmt.Sprintf("当前时间：%s\n今晚的晚霞有些烧哦，天空中会有漂亮的橙红色，记得去享受一下这份美丽！⛅", time.Now().Format("2006-01-02 15:04:05"))
	case "大烧":
		message = fmt.Sprintf("当前时间：%s\n今晚的晚霞真的是大烧，天边的火红色太美了，赶快去看吧！🔥", time.Now().Format("2006-01-02 15:04:05"))
	default:
		message = fmt.Sprintf("当前时间：%s\n天气数据获取失败，请稍后再试。", time.Now().Format("2006-01-02 15:04:05"))
	}
	return message
}

// 发送消息到企业微信 Webhook
func sendWxMsg(message string) error {
	// 直接在代码中写死 Webhook URL
	webhookURL := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=c67b1bd4-823f-459c-8940-8a73e4499172"

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

// 定时任务：每天 5:30 PM 发送天气和晚霞消息
func scheduleWeatherPush() {
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
		// 获取天气数据并推送消息
		weather, err := getWeather()
		if err != nil {
			log.Fatalf("获取天气数据失败: %v", err)
		}

		// 获取当前的云层覆盖度（假设我们取未来1小时的平均值）
		avgCloudCover := weather.Hourly.Cloudcover[0]

		// 判断晚霞情况
		quality := getSunsetQuality(avgCloudCover, weather.Current.Sunset)

		// 生成更人性化的消息内容
		message := generateMessage(quality)

		// 发送消息到企业微信
		if err := sendWxMsg(message); err != nil {
			log.Fatalf("发送消息失败: %v", err)
		}

		log.Println("消息发送成功！")

		// 每天执行一次
		time.Sleep(24 * time.Hour) // 间隔24小时再次发送
	}
}

// 主动触发推送：通过 HTTP 请求触发
func triggerPushHandler(w http.ResponseWriter, r *http.Request) {
	// 获取天气数据并推送消息
	weather, err := getWeather()
	if err != nil {
		http.Error(w, fmt.Sprintf("获取天气数据失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 获取当前的云层覆盖度（假设我们取未来1小时的平均值）
	avgCloudCover := weather.Hourly.Cloudcover[0]

	// 判断晚霞情况
	quality := getSunsetQuality(avgCloudCover, weather.Current.Sunset)

	// 生成更人性化的消息内容
	message := generateMessage(quality)

	// 发送消息到企业微信
	if err := sendWxMsg(message); err != nil {
		http.Error(w, fmt.Sprintf("发送消息失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.Write([]byte("消息发送成功！"))
}

func main() {
	// 启动定时任务
	go scheduleWeatherPush()

	// 启动 HTTP 服务，允许主动触发发送消息
	http.HandleFunc("/trigger-push", triggerPushHandler)
	log.Println("HTTP 服务已启动，监听端口 8080...")

	// 启动 HTTP 服务
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("启动 HTTP 服务失败: %v", err)
	}
}
