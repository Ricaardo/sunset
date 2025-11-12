# 🚀 快速开始

5分钟内完成部署！

## 前置要求

- Docker 已安装并运行
- 企业微信 Webhook URL（在企业微信群聊中获取）

## 三步部署

### 1️⃣ 克隆代码

```bash
git clone <项目仓库地址>
cd sunset
```

### 2️⃣ 配置 Webhook

编辑 `docker-compose.yml`，修改第 13 行：

```yaml
- WEBHOOK_URL=https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=你的密钥
```

**可选配置：**

```yaml
# 选择触发模式
- USE_SUNSET_TIME=true              # 使用日落时间触发
- SUNSET_ADVANCE_MINUTES=30         # 日落前30分钟推送

# 或者使用固定时间
- USE_SUNSET_TIME=false
- SCHEDULE_HOUR=17                  # 17:30 推送
- SCHEDULE_MINUTE=30

# 修改城市
- CITY=北京市-北京
- LATITUDE=39.9042
- LONGITUDE=116.4074
```

### 3️⃣ 一键部署

```bash
./deploy.sh
```

搞定！🎉

## 验证部署

部署成功后，你会看到：

```
✅ 一键部署完成！

服务信息:
  - 容器名称: sunset-app
  - 访问端口: 8080

可用的 API 接口:
  - 健康检查:   http://localhost:8080/health
  - 查询配置:   http://localhost:8080/config
  - 主动触发:   http://localhost:8080/trigger-push
  - 日落时间:   http://localhost:8080/sunset-time
```

## 测试推送

```bash
# 方式1: 使用管理脚本
./manage.sh test

# 方式2: 直接调用 API
curl -X POST http://localhost:8080/trigger-push
```

如果配置正确，你会在企业微信群里收到火烧云预报消息！

## 常用操作

```bash
# 查看实时日志
./manage.sh logs

# 查看当前配置
./manage.sh config

# 重启服务
./manage.sh restart

# 停止服务
./manage.sh stop
```

## 日常使用

部署后，服务会自动运行：

**固定时间模式：**
- 每天在设定时间（如 17:30）自动推送

**日落时间模式：**
- 每天在日落前 N 分钟（如 30 分钟）自动推送
- 时间随季节自动调整

你也可以随时手动触发：
```bash
curl -X POST http://localhost:8080/trigger-push
```

## 故障排查

### 问题1: 部署脚本报错 "Docker 未运行"

**解决：** 启动 Docker Desktop

### 问题2: 端口 8080 被占用

**解决：** 修改 `docker-compose.yml` 中的端口映射：
```yaml
ports:
  - "9090:8080"  # 改为使用 9090 端口
```

然后在环境变量中添加：
```yaml
- PORT=8080  # 容器内部端口保持 8080
```

访问时使用：`http://localhost:9090`

### 问题3: 没有收到消息

**检查步骤：**

1. 验证 Webhook URL 是否正确
   ```bash
   ./manage.sh config
   ```

2. 查看服务日志
   ```bash
   ./manage.sh logs
   ```

3. 手动测试推送
   ```bash
   ./manage.sh test
   ```

4. 确认企业微信机器人未被禁用

### 问题4: 时间不准确

**解决：**
- 服务已配置使用北京时间（UTC+8）
- 容器内时区设置为 `Asia/Shanghai`
- 可以通过 API 验证：
  ```bash
  curl http://localhost:8080/config
  ```

## 更新服务

当代码有更新时：

```bash
# 拉取最新代码
git pull

# 更新服务
./manage.sh update
```

## 卸载

```bash
# 停止并删除容器
docker compose down

# 删除镜像（可选）
docker rmi sunset-app

# 删除代码
cd .. && rm -rf sunset
```

## 获取帮助

- 查看完整文档：[README.md](README.md)
- 查看管理脚本帮助：`./manage.sh help`
- 问题反馈：提交 Issue

---

**祝你每天都能看到美丽的晚霞！** 🌅
