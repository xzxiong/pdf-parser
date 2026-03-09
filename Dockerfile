# 构建阶段
FROM golang:1.24-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git ca-certificates

# 设置工作目录
WORKDIR /build

# 复制 go.mod 和 go.sum 并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建二进制文件
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-w -s" -o pdf-parser .

# 运行阶段
FROM ubuntu:22.04

# 安装运行时依赖
RUN apk --no-cache add ca-certificates

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/pdf-parser .

# 更改所有权
RUN chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 设置入口点
ENTRYPOINT ["/app/pdf-parser", "--sleep"]
