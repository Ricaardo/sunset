# 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制依赖文件
COPY go.mod .

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sunset .

# 运行阶段
FROM alpine:latest

# 安装时区数据和 CA 证书（用于 HTTPS 请求）
RUN echo "http://mirrors.aliyun.com/alpine/v3.18/main" > /etc/apk/repositories && \
    echo "http://mirrors.aliyun.com/alpine/v3.18/community" >> /etc/apk/repositories && \
    apk update && \
    apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/sunset .

# 设置时区环境变量（默认为上海时区）
ENV TZ=Asia/Shanghai

# 暴露端口（程序监听8080端口）
EXPOSE 8080

# 运行应用
CMD ["./sunset"]