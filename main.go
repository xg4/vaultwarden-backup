package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/xg4/vaultwarden-backup/internal/app"
	"github.com/xg4/vaultwarden-backup/internal/config"
)

func bootstrap(cfg *config.Config) {
	// 记录开始时间
	startTime := time.Now()
	slog.Info("==================== 备份开始 ====================")

	// 创建并运行备份应用
	backupApp := app.New(cfg)
	if err := backupApp.Run(); err != nil {
		slog.Error(fmt.Sprintf("备份过程中发生错误: %v", err))
		slog.Error("==================== 备份失败 ====================")
		os.Exit(1)
	}

	// 记录结束和用时
	duration := time.Since(startTime).Seconds()
	slog.Info(fmt.Sprintf("用时: %.2f 秒", duration))
	slog.Info("==================== 备份完成 ====================")
}

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
		slog.Error(fmt.Sprintf("配置加载失败: %v", err))
		os.Exit(1)
	}

	// 打印关键配置
	slog.Info("-------------------- 环境变量 --------------------")
	slog.Info(fmt.Sprintf("备份目录 (BACKUP_DIR): %s", cfg.BackupDir))
	slog.Info(fmt.Sprintf("数据目录 (DATA_DIR): %s", cfg.DataDir))
	slog.Info(fmt.Sprintf("备份保留天数 (RETENTION_DAYS): %d", cfg.RetentionDays))
	slog.Info(fmt.Sprintf("最大并发数 (MAX_CONCURRENCY): %d", cfg.MaxConcurrency))
	slog.Info(fmt.Sprintf("备份间隔 (BACKUP_INTERVAL): %v", cfg.BackupInterval))
	slog.Info("--------------------------------------------------")

	bootstrap(cfg)

	ticker := time.NewTicker(cfg.BackupInterval)
	defer ticker.Stop()

	for range ticker.C {
		bootstrap(cfg)
	}
}
