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

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/sunset .

# 暴露端口（程序监听8080端口）
EXPOSE 8080

# 运行应用
CMD ["./sunset"]