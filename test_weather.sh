#!/bin/bash

# 天气功能测试脚本

echo "=========================================="
echo "  测试 Open-Meteo 天气API"
echo "=========================================="
echo ""

# 测试不同城市的天气
cities=(
    "上海:31.2304:121.4737"
    "北京:39.9042:116.4074"
    "广州:23.1291:113.2644"
)

for city_info in "${cities[@]}"; do
    IFS=':' read -r city lat lon <<< "$city_info"

    echo "📍 测试城市: $city (纬度: $lat, 经度: $lon)"
    echo "正在获取天气数据..."

    # 调用 Open-Meteo API
    response=$(curl -s "https://api.open-meteo.com/v1/forecast?latitude=$lat&longitude=$lon&current=temperature_2m,precipitation,cloud_cover,weather_code&timezone=Asia/Shanghai")

    # 解析响应
    if [ $? -eq 0 ]; then
        temperature=$(echo "$response" | grep -o '"temperature_2m":[0-9.]*' | cut -d':' -f2)
        precipitation=$(echo "$response" | grep -o '"precipitation":[0-9.]*' | cut -d':' -f2)
        cloud_cover=$(echo "$response" | grep -o '"cloud_cover":[0-9]*' | cut -d':' -f2)
        weather_code=$(echo "$response" | grep -o '"weather_code":[0-9]*' | cut -d':' -f2)

        echo "  温度: ${temperature}°C"
        echo "  降水量: ${precipitation}mm"
        echo "  云量: ${cloud_cover}%"
        echo "  天气代码: $weather_code"

        # 判断天气影响
        if (( $(echo "$precipitation > 5" | bc -l 2>/dev/null || echo 0) )); then
            echo "  ⚠️  影响等级: 严重 (大雨)"
        elif (( $(echo "$precipitation > 0" | bc -l 2>/dev/null || echo 0) )) || [ "$cloud_cover" -gt 85 ]; then
            echo "  🌧️  影响等级: 中等 (降水或云量高)"
        elif [ "$cloud_cover" -gt 70 ]; then
            echo "  ⛅  影响等级: 轻微 (多云)"
        else
            echo "  ✅  影响等级: 无影响 (晴朗)"
        fi
    else
        echo "  ❌ 获取天气数据失败"
    fi

    echo ""
done

echo "=========================================="
echo "  测试完成！"
echo "=========================================="
echo ""
echo "说明："
echo "- Open-Meteo API 完全免费，无需配置"
echo "- 系统会根据天气自动调整火烧云预测等级"
echo "- 天气API失败时会自动降级，不影响主功能"
