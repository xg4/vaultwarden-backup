# Vaultwarden 备份工具

这是一个用于自动备份 Vaultwarden 数据的工具。它能够备份数据库、配置文件、RSA 密钥、附件和发送文件等重要数据，并将其打包成加密的压缩文件。

## 功能特点

- 自动备份 Vaultwarden 的所有重要数据：
  - SQLite 数据库
  - 配置文件（config.json 和 .env）
  - RSA 密钥
  - 附件文件
  - 发送文件
- 使用 AES-256-CBC 加密备份文件
- 自动生成随机密码
- 支持 Docker 部署
- 支持定时备份（默认每 6 小时执行一次）

## 使用方法

### Docker 部署

1. 拉取镜像：

```bash
docker pull ghcr.io/xg4/vaultwarden-backup
```

2. 运行容器：

```bash
docker run -d \
  --name vaultwarden-backup \
  -v /path/to/vaultwarden/data:/data \
  -v /path/to/backup:/backups \
  -e ZIP_PASSWORD=your-password \
  --restart unless-stopped \
  ghcr.io/xg4/vaultwarden-backup
```

### 环境变量

- `DATA_DIR`: Vaultwarden 数据目录路径（默认：`/data`）
- `BACKUP_DIR`: 备份文件存储路径（默认：`/backups/backup`）
- `ZIP_PASSWORD`: 备份文件加密密码（**必需**）。脚本在执行前会检查此环境变量是否设置，如果未设置则会报错并退出。
- `RETENTION_DAYS`: 备份文件保留天数（默认：`30`）。脚本会清理超过此天数的旧备份文件。设置为 `0` 则不清理。

### 手动执行备份

```bash
./backup.sh
```

## 备份文件说明

备份文件将以 `backup_YYYYMMDD_HHMMSS.tar.gz` 的格式命名，并使用您在 `ZIP_PASSWORD` 环境变量中设置的密码通过 AES-256-CBC 加密。

### 还原备份

您可以使用以下命令来还原备份文件：

```bash
openssl enc -d -aes-256-cbc -salt -pbkdf2 -pass pass:your-password -in backup_YYYYMMDD_HHMMSS.tar.gz | tar xz -C .
```

**注意：**

- 将 `backup_YYYYMMDD_HHMMSS.tar.gz` 替换为您的实际备份文件名。
- 将 `your-password` 替换为您的备份加密密码。
- `-C .` 表示将文件解压到当前目录，您可以根据需要修改目标目录。

## 依赖

- sqlite3
- tar
- openssl
- bash

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。
