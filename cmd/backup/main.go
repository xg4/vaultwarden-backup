package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/xg4/vaultwarden-backup/internal/app"
	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/logger"
)

func main() {
	// 初始化日志记录器
	logger.Setup()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		slog.Error("🚨 配置加载失败", "error", err)
		os.Exit(1)
	}

	// 显示关键配置信息
	slog.Info("🚀 启动备份服务",
		"DATA_DIR", cfg.DataDir,
		"BACKUP_DIR", cfg.BackupDir,
		"BACKUP_INTERVAL", cfg.BackupInterval,
		"RETENTION_DAYS", cfg.RetentionDays)

	// 设置优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 创建备份应用
	backupApp := app.New(cfg)

	// 执行初始备份
	slog.Info("📦 执行初始备份")
	if err := backupApp.Run(); err != nil {
		slog.Error("🚨 初始备份失败", "error", err)
	}

	// 启动定时备份
	ticker := time.NewTicker(cfg.BackupInterval)
	defer ticker.Stop()

	slog.Info("⏰ 定时备份已启动", "interval", cfg.BackupInterval)

	// 用于跟踪正在进行的备份
	var backupInProgress bool
	var backupMutex sync.Mutex

	for {
		select {
		case <-ctx.Done():
			slog.Info("🛑 收到停止信号，正在关闭...")
			return
		case <-sigChan:
			slog.Info("🔄 收到系统信号，正在优雅关闭...")

			// 检查是否有备份正在进行
			backupMutex.Lock()
			if backupInProgress {
				slog.Info("⏳ 等待当前备份任务完成...")
				backupMutex.Unlock()

				// 等待备份完成，最多等待10 秒
				timeout := time.NewTimer(1 * time.Second)
				defer timeout.Stop()

				for {
					time.Sleep(time.Second)
					backupMutex.Lock()
					if !backupInProgress {
						backupMutex.Unlock()
						break
					}
					backupMutex.Unlock()

					select {
					case <-timeout.C:
						slog.Warn("⚠️ 等待备份完成超时，强制退出")
						cancel()
						return
					default:
					}
				}
			} else {
				backupMutex.Unlock()
			}

			slog.Info("✅ 备份任务已完成，安全退出")
			cancel()
			return
		case <-ticker.C:
			// 检查是否已经有备份在进行
			backupMutex.Lock()
			if backupInProgress {
				slog.Debug("⏭️ 跳过定时备份，上一个备份仍在进行中")
				backupMutex.Unlock()
				continue
			}
			backupInProgress = true
			backupMutex.Unlock()

			slog.Debug("🔄 开始定时备份")
			go func() {
				defer func() {
					backupMutex.Lock()
					backupInProgress = false
					backupMutex.Unlock()
				}()

				if err := backupApp.Run(); err != nil {
					slog.Error("🚨 定时备份失败", "error", err)
				}
			}()
		}
	}
}
