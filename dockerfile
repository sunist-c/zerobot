# 使用官方的 Go 镜像作为构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制项目的源代码
COPY . .

# 下载依赖
RUN go mod download

# 编译 Go 应用
RUN go build -o zerobot-plugin .

# 使用一个更小的镜像作为运行阶段
FROM alpine:latest

# 创建一个目录来存放应用程序
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/zerobot-plugin .

# 复制配置文件
COPY config ./config

# 设置环境变量，如果有需要的话
# ENV VAR_NAME value

# 运行应用程序
CMD ["./zerobot-plugin", "-c", "config/config.json", "-m", "config/config.yaml", "-g", "config/public.yaml"]