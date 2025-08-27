# 使用支持 Go 1.23 的官方 Go 语言镜像作为基础镜像
FROM golang:1.23-alpine as build

# 设置工作目录
WORKDIR /app

# 将当前目录内容复制到工作目录
COPY . .

# 下载依赖并编译 Go 项目
RUN go mod tidy
RUN go build -o sunset .

# 使用 Alpine 镜像作为运行时环境
FROM alpine:latest

# 安装必要的依赖
RUN apk --no-cache add ca-certificates

# 将编译好的程序复制到运行时镜像
COPY --from=build /app/sunset /usr/local/bin/sunset

# 设置容器启动命令
ENTRYPOINT ["/usr/local/bin/sunset"]

# 设置容器默认运行时
CMD []
