package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/xg4/vaultwarden-backup/internal/app"
	"github.com/xg4/vaultwarden-backup/internal/config"
)

func main() {
	// 初始化日志记录器
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)

	slog.SetDefault(slog.New(handler))

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		slog.Error("配置加载失败", "error", err)
		os.Exit(1)
	}

	// 显示关键配置信息
	slog.Info("启动备份服务",
		"data_dir", cfg.DataDir,
		"backup_dir", cfg.BackupDir,
		"interval", cfg.BackupInterval,
		"retention_days", cfg.RetentionDays)

	// 创建并运行备份应用
	backupApp := app.New(cfg)
	if err := backupApp.Run(); err != nil {
		os.Exit(1)
	}

	ticker := time.NewTicker(cfg.BackupInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := backupApp.Run(); err != nil {
			os.Exit(1)
		}
	}
}
