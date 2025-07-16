# Vaultwarden Backup Tool

English | [中文](README.md)

A simple and easy-to-use automatic backup tool for [Vaultwarden](https://github.com/dani-garcia/vaultwarden), supporting scheduled backups, encrypted storage, and one-click recovery.

## ✨ Features

- 🔄 **Automatic Backup** - Scheduled backup of all important data (database, config, attachments, etc.)
- 🔐 **Secure Encryption** - Encrypt backup files using AES-256-GCM algorithm
- 🐳 **Containerized** - Ready-to-use Docker image
- 🧹 **Auto Cleanup** - Automatically delete expired backup files
- ⚡ **High Performance** - Parallel execution of backup tasks for faster speed

## 🚀 Quick Start

### One-Click Deployment

```bash
# Pull the image
docker pull ghcr.io/xg4/vaultwarden-backup:latest

# Start backup service (please modify paths and password)
docker run -d \
  --name vaultwarden-backup \
  -v /path/to/vaultwarden/data:/data \
  -v /path/to/backups:/backups \
  -e PASSWORD=your-strong-password \
  -e RETENTION_DAYS=30 \
  --restart unless-stopped \
  ghcr.io/xg4/vaultwarden-backup:latest
```

> ⚠️ **Important**: Please replace `/path/to/vaultwarden/data` and `/path/to/backups` with actual paths, and set a strong password

## ⚙️ Configuration Options

| Environment Variable | Default Value | Description                                                       |
| -------------------- | ------------- | ----------------------------------------------------------------- |
| `PASSWORD`           | _Required_    | 🔑 Backup file encryption password (please set a strong password) |
| `BACKUP_INTERVAL`    | `1h`          | ⏰ Backup interval time (supports `s`/`m`/`h`, e.g., `6h`)        |
| `RETENTION_DAYS`     | `30`          | 🗂️ Backup retention days (set to `0` to disable cleanup)          |
| `DATA_DIR`           | `/data`       | 📁 Vaultwarden data directory path                                |
| `BACKUP_DIR`         | `/backups`    | 💾 Backup file storage path                                       |

## 📋 Common Operations

### Manual Backup

```bash
docker exec vaultwarden-backup vaultb
```

### Restore Backup

```bash
# Decrypt and restore backup file
docker run --rm -it \
  -v /path/to/backups:/backups \
  ghcr.io/xg4/vaultwarden-backup vaultr \
  -i /backups/vault_20240101_120000.tar.gz \
  -o /backups/restored \
  -p your_password
```

### View Logs

```bash
docker logs vaultwarden-backup
```

## 📝 Backup Information

- **File Format**: `vault_YYYYMMDD_HHMMSS.tar.gz`
- **Encryption Method**: AES-256-GCM algorithm
- **Backup Content**: Database, configuration files, RSA keys, attachments, send files

## 📄 License

MIT License - See [LICENSE](LICENSE) file for details
