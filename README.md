# Sunset 部署指南

## 项目介绍
Sunset 是一个自动推送日落晚霞预报的服务，可定时向企业微信发送晚霞质量等级预报。

## 部署方式

### Docker Compose 部署（推荐）
新版本 Docker 已内置 Docker Compose，无需额外安装。

#### 步骤 1: 准备环境
确保已安装 Docker 环境（推荐版本 20.10+）

#### 步骤 2: 克隆代码库
```bash
git clone <项目仓库地址>
cd sunset
```

#### 步骤 3: 配置企业微信 Webhook（可选）
修改 `main.go` 文件中的企业微信 Webhook 地址：
```go
// 在 sendWxMarkdownMsg 函数中修改
webhookURL := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=你的企业微信Webhook密钥"
```

#### 步骤 4: 使用 Docker Compose 部署
```bash
# 构建镜像并启动服务（首次运行或代码更新后使用）
docker compose up -d --build

# 查看服务状态
docker compose ps

# 查看日志输出
docker compose logs -f sunset
```

#### 常用命令
```bash
# 停止服务
docker compose down

# 重启服务
docker compose restart

# 进入容器内部
docker compose exec sunset sh

# 查看服务详细信息
docker compose inspect sunset
```

## 功能验证
服务启动后，可以通过以下方式验证：

1. 访问 `http://localhost:8080/trigger-push` 主动触发一次消息推送
2. 检查企业微信是否收到消息
3. 服务会在每天 17:30 自动推送当日晚霞预报

## 注意事项
- 时区已配置为 Asia/Shanghai，确保定时任务时间准确
- 若需修改端口映射，可编辑 `docker-compose.yml` 文件中的 `ports` 配置
- 代码中的城市配置默认为上海市，如需修改可调整 `getSunsetData` 函数中的请求 URL
- 容器会在意外停止后自动重启（除非手动停止）

## 依赖说明
- Go 1.23 及以上版本
- Docker 20.10+（用于容器化部署）