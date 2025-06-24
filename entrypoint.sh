#!/bin/sh

# 执行一次备份
/usr/local/bin/vault-backup

# 启动 cron 守护进程
exec crond -f