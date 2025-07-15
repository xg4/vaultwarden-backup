package app

import (
	"log/slog"
	"os"
	"time"

	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/engine"
	"github.com/xg4/vaultwarden-backup/internal/tasks"
)

// App 封装了备份应用的状态和依赖
type App struct {
	cfg *config.Config
}

// New 创建一个新的 App 实例
func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

// Run 执行完整的备份和清理流程
func (a *App) Run() error {
	startTime := time.Now()
	defer func(t time.Time) {
		duration := time.Since(t)
		slog.Info("备份用时", "duration", duration)
	}(startTime)

	timestamp := startTime.Format("20060102_150405")
	slog.Info("备份时间", "timestamp", timestamp)
	slog.Info("==================== 备份开始 ====================")

	s := engine.New(a.cfg)
	// 1. 运行前检查
	s.Register(
		&tasks.CheckDataDir{},
		&tasks.CheckDiskSpace{},
		&tasks.CreateBackupTmpDir{},
	)

	// 2. 初始化备份环境

	// 使用 defer 确保临时目录总能被清理
	defer os.RemoveAll(a.cfg.BackupTmpDir)

	// 3. 执行所有备份任务
	s.Register(
		&tasks.DatabaseTask{},
		&tasks.RSATask{},
		&tasks.CopyTask{Path: ".env"},
		&tasks.CopyTask{Path: "config.json"},
		&tasks.CopyTask{Path: "attachments"},
		&tasks.CopyTask{Path: "sends"},
	)

	// 4. 打包压缩和加密
	s.Register(&tasks.ArchiveTask{
		Timestamp: timestamp,
	})

	// 5. 清理旧备份
	s.Register(&tasks.CleanupTask{})
	if err := s.Start(); err != nil {
		slog.Error("备份过程中发生错误", "error", err)
		slog.Error("==================== 备份失败 ====================")
		return err
	}

	slog.Info("==================== 备份完成 ====================")
	return nil
}
