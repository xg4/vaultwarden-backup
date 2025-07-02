# 构建阶段
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache gcc musl-dev sqlite-dev

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建静态二进制文件
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o vault-backup .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o backup ./cmd/backup
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o restore ./cmd/restore

# 运行阶段
FROM alpine:latest

RUN apk add --no-cache \
    sqlite

# 复制构建的二进制文件
COPY --from=builder /app/vault-backup /usr/local/bin/vault-backup
COPY --from=builder /app/backup /usr/local/bin/backup
COPY --from=builder /app/restore /usr/local/bin/restore

# 复制入口脚本
COPY entrypoint.sh /

RUN chmod +x /entrypoint.sh && \
    chmod +x /usr/local/bin/vault-backup && \
    chmod +x /usr/local/bin/backup && \
    chmod +x /usr/local/bin/restore

ENTRYPOINT ["/entrypoint.sh"]