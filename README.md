# Vaultwarden 备份工具

这是一个用 Go 语言编写的、用于自动备份 [Vaultwarden](https://github.com/dani-garcia/vaultwarden) 数据的工具。它被设计为在 Docker 容器中运行，能够备份数据库、配置文件、RSA 密钥、附件和发送文件等重要数据，并将其打包成加密的压缩文件。

## 功能特点

- **全量备份**: 自动备份 Vaultwarden 的所有重要数据：
  - SQLite 数据库 (`db.sqlite3`)
  - 配置文件 (`config.json` 和 `.env`)
  - RSA 密钥 (`rsa_key*`, `rsa_key.pem`, `rsa_key.pub.pem`)
  - 附件目录 (`attachments`)
  - 发送文件目录 (`sends`)
- **安全加密**: 使用 **AES-256-GCM** 算法加密备份文件，这是一种经过身份验证的加密模式，可提供强大的安全保障。
- **容器化部署**: 通过 Docker 镜像提供，易于部署和管理。
- **定时任务**: 使用 Go Ticker 实现定时任务，默认为每 6 小时执行一次备份。
- **自动清理**: 支持设置备份保留天数，自动清理过期的备份文件。
- **并发执行**: 支持并发执行备份任务，提升备份效率。

## 使用方法

### Docker 部署

1.  **拉取镜像**:
    从 Github Packages 拉取最新的 Docker 镜像。

    ```bash
    docker pull ghcr.io/xg4/vaultwarden-backup:latest
    ```

2.  **运行容器**:
    创建并运行 `vaultwarden-backup` 容器。请确保将 `/path/to/vaultwarden/data` 替换为你实际的 Vaultwarden 数据目录路径，并将 `/path/to/backups` 替换为你希望存放备份文件的目录。

    ```bash
    docker run -d \
      --name vaultwarden-backup \
      -v /path/to/vaultwarden/data:/data \
      -v /path/to/backups:/backups \
      -e PASSWORD=your-password \
      -e RETENTION_DAYS=30 \
      --restart unless-stopped \
      ghcr.io/xg4/vaultwarden-backup:latest
    ```

### 环境变量

通过配置环境变量来自定义备份行为：

- `DATA_DIR`: Vaultwarden 数据目录在容器内的路径 (默认: `/data`)
- `BACKUP_DIR`: 备份文件的存储路径 (默认: `/backups`)
- `PASSWORD`: **(必需)** 用于加密备份文件的密码。**请务必设置一个强密码**。如果未设置，程序将报错并退出。
- `RETENTION_DAYS`: 备份文件保留天数 (默认: `30`)。脚本会清理超过此天数的旧备份文件。设置为 `0` 则禁用自动清理。
- `BACKUP_INTERVAL`: 定时备份的间隔时间 (默认: `1h`),支持 `s`, `m`, `h` 单位,例如 `8h`, `12h`。

## 手动执行

除了自动定时备份，你也可以手动在容器内执行备份或恢复命令。

### 手动备份

```bash
docker exec vaultwarden-backup vaultb
```

### 手动恢复

项目提供了 `vaultr` 命令用于解密并解压备份文件。该命令支持以下参数：

- `-i`, `-input`: **(必需)** 输入的加密备份文件路径。
- `-o`, `-output`: **(必需)** 解密后文件的输出目录路径。
- `-p`, `-password`: **(必需)** 用于解密的密码。
- `-v`, `-verbose`: 启用详细输出模式。
- `-h`, `-help`: 显示帮助信息。

```bash
docker run --rm -it \
  -v /path/to/backups:/backups \
  ghcr.io/xg4/vaultwarden-backup vaultr \
  -i /backups/vault_YYYYMMDD_HHMMSS.tar.gz \
  -o /backups/restored \
  -p your_password
```

## 备份文件

- 备份文件以 `vault_YYYYMMDD_HHMMSS.tar.gz` 的格式命名。
- 文件使用 `PASSWORD` 环境变量中设置的密码通过 **AES-256-GCM** 加密。

## 依赖

本工具在 Docker 容器中运行，运行时仅依赖于 `sqlite`，无需其他外部工具。

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。
