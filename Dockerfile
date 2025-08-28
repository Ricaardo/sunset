FROM golang:1.23-alpine AS builder

WORKDIR /app

# 复制依赖文件并下载依赖
COPY go.mod .
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -o sunset .

# 使用轻量镜像运行
FROM alpine:latest

WORKDIR /root/

# 从构建阶段复制编译好的应用
COPY --from=builder /app/sunset .

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./sunset"]