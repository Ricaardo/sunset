import requests
import datetime

def get_weather_data():
    """获取上海明天的气象数据"""
    url = "https://api.open-meteo.com/v1/forecast"
    params = {
        "latitude": 31.23,
        "longitude": 121.47,
        "hourly": "cloudcover,cloudcover_low,cloudcover_mid,cloudcover_high",
        "daily": "sunset",
        "timezone": "Asia/Shanghai",
        "forecast_days": 1
    }
    
    response = requests.get(url, params=params)
    if response.status_code == 200:
        return response.json()
    else:
        print(f"天气数据请求失败: {response.text}")
        return None

def get_aod_proxy():
    """通过PM2.5间接获取气溶胶光学厚度代理数据"""
    url = "https://api.openaq.org/v2/latest"
    params = {
        "city": "Shanghai",
        "parameter": "pm25",
        "limit": 1
    }
    
    try:
        response = requests.get(url, params=params)
        if response.status_code == 200:
            data = response.json()
            if data["results"]:
                pm25 = data["results"][0]["measurements"][0]["value"]
                return pm25
    except:
        pass
    return 50  # 默认中等污染

def calculate_firecloud_score(weather_data, pm25):
    """计算火烧云鲜艳度得分"""
    if not weather_data:
        return 0
    
    # 提取日落时间（明天）
    sunset = weather_data["daily"]["sunset"][0]
    sunset_hour = int(sunset.split("T")[1].split(":")[0])
    
    # 获取日落前后3小时的云量数据（18-21点，假设日落在18-19点）
    hourly_data = weather_data["hourly"]
    target_hours = [18, 19, 20, 21]
    cloud_data = []
    
    for i, time in enumerate(hourly_data["time"]):
        hour = int(time.split("T")[1].split(":")[0])
        if hour in target_hours:
            cloud_data.append({
                "total": hourly_data["cloudcover"][i],
                "low": hourly_data["cloudcover_low"][i],
                "mid": hourly_data["cloudcover_mid"][i],
                "high": hourly_data["cloudcover_high"][i]
            })
    
    if not cloud_data:
        return 0
    
    # 1. 持续时间得分（0-1）
    valid_hours = sum(1 for c in cloud_data if 30 <= c["total"] <= 80)
    duration_score = min(valid_hours / 3, 1.0)
    
    # 2. 云量结构得分（中高云为主，低云少）（0-1）
    mid_high_ratio = sum((c["mid"] + c["high"]) / max(c["total"], 1) for c in cloud_data) / len(cloud_data)
    cloud_structure_score = mid_high_ratio * 0.8 + (1 - sum(c["low"]/100 for c in cloud_data)/len(cloud_data)) * 0.2
    
    # 3. 天空占比得分（0-1）
    avg_cloudcover = sum(c["total"] for c in cloud_data) / len(cloud_data)
    area_score = min(avg_cloudcover / 70, 1.0) if avg_cloudcover >= 30 else avg_cloudcover / 30
    
    # 4. 大气通透度得分（基于PM2.5）（0-1）
    if pm25 <= 35:
        aod_score = 0.9
    elif pm25 <= 75:
        aod_score = 0.5
    else:
        aod_score = 0.2
    
    # 5. 综合得分（加权计算）
    total_score = (
        duration_score * 0.2 +
        cloud_structure_score * 0.3 +
        area_score * 0.2 +
        aod_score * 0.3
    )
    
    # 映射到鲜艳度等级范围（0-2.5）
    return min(total_score * 2.5, 2.5)

def get_firecloud_level(score):
    """根据得分返回火烧云等级描述"""
    if score < 0.05:
        return "微微烧，或者火烧云云况不典型没有预报出来"
    elif score < 0.2:
        return "小烧，大气很通透的情况下才会比较好看"
    elif score < 0.4:
        return "小烧到中等烧"
    elif score < 0.6:
        return "中等烧，比较值得看的火烧云"
    elif score < 0.8:
        return "中等烧到大烧程度的火烧云"
    elif score < 1.0:
        return "不是很完美的大烧火烧云"
    elif score < 1.5:
        return "典型的火烧云大烧"
    elif score < 2.0:
        return "优质大烧，火烧云范围广、云量大、颜色明亮"
    else:
        return "世纪大烧，范围广、接近满云量、颜色鲜艳、持续时间长"

if __name__ == "__main__":
    print("上海明天火烧云预报分析...")
    weather_data = get_weather_data()
    pm25 = get_aod_proxy()
    score = calculate_firecloud_score(weather_data, pm25)
    level = get_firecloud_level(score)
    
    print(f"\n鲜艳度得分: {score:.2f}")
    print(f"等级评价: {level}")
    print(f"PM2.5浓度（参考）: {pm25}μg/m³")
