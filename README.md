# Vaultwarden 备份工具

[English](README_EN.md) | 中文

一个简单易用的 [Vaultwarden](https://github.com/dani-garcia/vaultwarden) 自动备份工具，支持定时备份、加密存储和一键恢复。

## ✨ 特性

- 🔄 **自动备份** - 定时备份所有重要数据（数据库、配置、附件等）
- 🔐 **安全加密** - 使用 AES-256-GCM 算法加密备份文件
- 🐳 **容器化** - 开箱即用的 Docker 镜像
- 🧹 **自动清理** - 自动删除过期备份文件
- ⚡ **高效并发** - 并行执行备份任务，速度更快

## 🚀 快速开始

### 一键部署

```bash
# 拉取镜像
docker pull ghcr.io/xg4/vaultwarden-backup:latest

# 启动备份服务（请修改路径和密码）
docker run -d \
  --name vaultwarden-backup \
  -v /path/to/vaultwarden/data:/data \
  -v /path/to/backups:/backups \
  -e PASSWORD=your-strong-password \
  -e BACKUP_INTERVAL=6h \
  -e PRUNE_BACKUPS_DAYS=7 \
  --restart unless-stopped \
  ghcr.io/xg4/vaultwarden-backup:latest
```

> ⚠️ **重要**: 请将 `/path/to/vaultwarden/data` 和 `/path/to/backups` 替换为实际路径，并设置强密码

## ⚙️ 配置选项

| 环境变量              | 默认值     | 说明                                                                |
| --------------------- | ---------- | ------------------------------------------------------------------- |
| `PASSWORD`            | _必需_     | 🔑 备份文件加密密码（请设置强密码）                                 |
| `BACKUP_INTERVAL`     | `6h`       | ⏰ 备份间隔时间（支持 `s`/`m`/`h`，如 `6h`）                        |
| `PRUNE_BACKUPS_DAYS`  | `30`       | 🗂️ 备份保留天数（设为 `0` 禁用清理）                                |
| `PRUNE_BACKUPS_COUNT` | `0`        | 🔢 保留的备份文件数量（设为 `0` 禁用，优先于 `PRUNE_BACKUPS_DAYS`） |
| `BACKUP_NAME`         | `vault`    | 📝 备份文件名前缀                                                   |
| `DATA_DIR`            | `/data`    | 📁 Vaultwarden 数据目录路径                                         |
| `BACKUP_DIR`          | `/backups` | 💾 备份文件存储路径                                                 |

## 📋 常用操作

### 手动备份

```bash
docker exec vaultwarden-backup vaultb
```

### 恢复备份

```bash
# 解密并恢复备份文件
docker run --rm -it \
  -v /path/to/backups:/backups \
  ghcr.io/xg4/vaultwarden-backup vaultr \
  -i /backups/vault_20240101_120000.tar.gz \
  -o /backups/restored \
  -p your_password
```

### 查看日志

```bash
docker logs vaultwarden-backup
```

## 📝 备份说明

- **文件格式**: `vault_YYYYMMDD_HHMMSS.tar.gz`
- **加密方式**: AES-256-GCM 算法
- **备份内容**: 数据库、配置文件、RSA 密钥、附件、发送文件

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件
