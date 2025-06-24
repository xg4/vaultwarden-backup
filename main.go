package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/xg4/vaultwarden-backup/internal/app"
	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/logger"
)

func main() {
	// 初始化日志记录器
	logger.InitLogger(slog.LevelInfo)
	l := logger.GetLogger()

	// 记录开始时间
	startTime := time.Now()
	l.Info("==================== 备份开始 ====================")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		l.Error(fmt.Sprintf("配置加载失败: %v", err))
		os.Exit(1)
	}

	// 创建并运行备份应用
	backupApp := app.New(cfg, l)
	if err := backupApp.Run(); err != nil {
		l.Error(fmt.Sprintf("备份过程中发生错误: %v", err))
		l.Error("==================== 备份失败 ====================")
		os.Exit(1)
	}

	// 记录结束和用时
	duration := time.Since(startTime).Seconds()
	l.Info(fmt.Sprintf("用时: %.2f 秒", duration))
	l.Info("==================== 备份完成 ====================")
}
