# Sunset 部署指南

## 项目介绍
Sunset 是一个自动推送日落晚霞预报的服务，可定时向企业微信发送晚霞质量等级预报。

> 📖 **快速开始**: 如果你想尽快部署，请查看 [快速开始指南](QUICKSTART.md)
> 🚀 **一键部署**: `./deploy.sh`

## 主要特性

- ⏰ **灵活的触发模式**：支持固定时间触发和日落时间自动触发
- 🌍 **北京时间支持**：所有时间均使用东八区（UTC+8）北京时间
- 🌆 **日落时间计算**：基于地理坐标自动计算每日日落时间
- 🔧 **环境变量配置**：支持通过环境变量灵活配置服务参数
- 🔌 **RESTful API**：提供多个 HTTP 接口用于主动触发和状态查询
- 📊 **配置查询**：可通过 API 查询当前配置和下次推送时间

## 环境变量配置

| 变量名 | 说明 | 默认值 | 示例 |
|--------|------|--------|------|
| `WEBHOOK_URL` | 企业微信 Webhook URL | - | `https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxxxx` |
| `CITY` | 城市名称 | `上海市-上海` | `北京市-北京` |
| `LATITUDE` | 纬度 | `31.2304` | `39.9042` |
| `LONGITUDE` | 经度 | `121.4737` | `116.4074` |
| `SCHEDULE_HOUR` | 定时推送小时（24小时制） | `17` | `18` |
| `SCHEDULE_MINUTE` | 定时推送分钟 | `30` | `0` |
| `USE_SUNSET_TIME` | 是否使用日落时间触发 | `false` | `true` |
| `SUNSET_ADVANCE_MINUTES` | 日落前提前多少分钟推送 | `30` | `45` |
| `PORT` | HTTP 服务端口 | `8080` | `9090` |

## 部署方式

### 🚀 一键部署（最简单）

```bash
# 1. 克隆代码库
git clone <项目仓库地址>
cd sunset

# 2. 修改配置（可选）
# 编辑 docker-compose.yml 修改 Webhook URL 等配置

# 3. 一键部署
./deploy.sh
```

就这么简单！脚本会自动：
- 检查 Docker 环境
- 清理旧容器
- 构建并启动服务
- 显示服务状态和配置

### 📋 快速管理命令

部署后可以使用管理脚本快速操作：

```bash
./manage.sh start    # 启动服务
./manage.sh stop     # 停止服务
./manage.sh restart  # 重启服务
./manage.sh logs     # 查看日志
./manage.sh status   # 查看状态
./manage.sh config   # 查看配置
./manage.sh test     # 测试推送
./manage.sh update   # 更新服务
```

### 🐳 Docker Compose 手动部署

如果想手动部署，可以按以下步骤操作：

#### 步骤 1: 准备环境
确保已安装 Docker 环境（推荐版本 20.10+）

#### 步骤 2: 克隆代码库
```bash
git clone <项目仓库地址>
cd sunset
```

#### 步骤 3: 配置环境变量
编辑 `docker-compose.yml` 文件中的环境变量：

```yaml
environment:
  - WEBHOOK_URL=https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=你的密钥
  - CITY=上海市-上海
  - LATITUDE=31.2304
  - LONGITUDE=121.4737
  - SCHEDULE_HOUR=17
  - SCHEDULE_MINUTE=30
  - USE_SUNSET_TIME=false  # 设置为 true 启用日落时间触发
  - SUNSET_ADVANCE_MINUTES=30  # 日落前提前30分钟推送
  - PORT=8080
```

#### 步骤 4: 构建并启动
```bash
docker compose up -d --build
```

#### 步骤 5: 验证服务
```bash
# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f sunset

# 测试 API
curl http://localhost:8080/health
```

### 本地运行

```bash
# 设置环境变量
export WEBHOOK_URL="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=你的密钥"
export USE_SUNSET_TIME=true

# 运行服务
go run main.go
```

## API 接口

服务启动后提供以下 HTTP 接口：

### 1. 主动触发推送
```bash
# GET 或 POST 请求
curl http://localhost:8080/trigger-push

# 响应示例
{
  "status": "success",
  "message": "消息发送成功",
  "timestamp": "2024-11-11 18:30:00"
}
```

### 2. 健康检查
```bash
curl http://localhost:8080/health

# 响应示例
{
  "status": "ok",
  "timestamp": "2024-11-11 18:30:00",
  "timezone": "Asia/Shanghai (UTC+8)"
}
```

### 3. 查询配置
```bash
curl http://localhost:8080/config

# 响应示例（固定时间模式）
{
  "city": "上海市-上海",
  "latitude": 31.2304,
  "longitude": 121.4737,
  "schedule_hour": 17,
  "schedule_minute": 30,
  "use_sunset_time": false,
  "timezone": "Asia/Shanghai (UTC+8)",
  "next_push_time": "2024-11-11 17:30:00",
  "current_time": "2024-11-11 16:00:00"
}

# 响应示例（日落时间模式）
{
  "city": "上海市-上海",
  "latitude": 31.2304,
  "longitude": 121.4737,
  "schedule_hour": 17,
  "schedule_minute": 30,
  "use_sunset_time": true,
  "sunset_advance_minutes": 30,
  "timezone": "Asia/Shanghai (UTC+8)",
  "next_push_time": "2024-11-11 17:15:00",
  "next_sunset_time": "2024-11-11 17:45:00",
  "current_time": "2024-11-11 16:00:00"
}
```

### 4. 查询日落时间
```bash
curl http://localhost:8080/sunset-time

# 响应示例
{
  "sunset_time": "2024-11-11 17:15:00",
  "current_time": "2024-11-11 16:00:00",
  "latitude": 31.2304,
  "longitude": 121.4737,
  "city": "上海市-上海"
}
```

## 触发模式说明

### 固定时间模式（默认）
- 设置 `USE_SUNSET_TIME=false`
- 每天在配置的固定时间（如 17:30）触发推送
- 适合想要在固定时间收到通知的场景

### 日落时间模式
- 设置 `USE_SUNSET_TIME=true`
- 根据地理坐标自动计算每日日落时间，并在日落前指定时间触发推送
- 通过 `SUNSET_ADVANCE_MINUTES` 设置提前时间（默认30分钟）
- 例如：日落时间为 18:00，提前30分钟则在 17:30 推送
- 日落时间会随季节变化自动调整
- 更符合实际观赏晚霞的最佳时间，给用户留出准备时间

## 主要城市坐标参考

| 城市 | CITY | LATITUDE | LONGITUDE |
|------|------|----------|-----------|
| 上海 | 上海市-上海 | 31.2304 | 121.4737 |
| 北京 | 北京市-北京 | 39.9042 | 116.4074 |
| 广州 | 广东省-广州 | 23.1291 | 113.2644 |
| 深圳 | 广东省-深圳 | 22.5431 | 114.0579 |
| 杭州 | 浙江省-杭州 | 30.2741 | 120.1551 |
| 成都 | 四川省-成都 | 30.5728 | 104.0668 |

## 注意事项

- 时区已配置为 Asia/Shanghai（UTC+8），确保定时任务时间准确
- 若需修改端口映射，可编辑 `docker-compose.yml` 文件中的 `ports` 配置
- 容器会在意外停止后自动重启（除非手动停止）
- 日落时间计算使用NOAA算法，精度约为 ±2 分钟（经过实测验证）
- 建议在使用日落时间模式前先通过 `/sunset-time` 接口验证计算结果

## 依赖说明

- Go 1.23 及以上版本
- Docker 20.10+（用于容器化部署）

## 故障排查

### 查看服务日志
```bash
docker compose logs -f sunset
```

### 测试推送功能
```bash
curl http://localhost:8080/trigger-push
```

### 验证配置
```bash
curl http://localhost:8080/config
```