# 多阶段构建 Dockerfile for CCany Go版本

# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建参数
ARG VERSION=dev
ARG BUILD_TIME=unknown

# 构建应用程序
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-extldflags '-static' -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
    -o ccany cmd/server/main.go

# 运行阶段
FROM alpine:latest

# 安装ca-certificates用于HTTPS请求
RUN apk --no-cache add ca-certificates tzdata wget

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/ccany .

# 复制web静态文件
COPY --from=builder /app/web ./web

# 复制脚本文件
COPY --from=builder /app/scripts ./scripts

# 创建数据目录和日志目录
RUN mkdir -p /app/data /app/logs && \
    chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8082

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# 启动应用程序
CMD ["./ccany"]
