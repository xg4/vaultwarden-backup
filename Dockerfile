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
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o vaultb ./cmd/backup
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o vaultr ./cmd/restore

# 运行阶段
FROM alpine:latest

RUN apk add --no-cache \
    sqlite

# 复制构建的二进制文件
COPY --from=builder /app/vaultb /usr/local/bin/vaultb
COPY --from=builder /app/vaultr /usr/local/bin/vaultr

RUN chmod +x /usr/local/bin/vaultb && \
    chmod +x /usr/local/bin/vaultr

CMD ["/usr/local/bin/vaultb"]