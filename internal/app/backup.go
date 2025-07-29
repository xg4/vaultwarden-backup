package app

import (
	"log/slog"
	"os"
	"time"

	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/scheduler"
	"github.com/xg4/vaultwarden-backup/internal/tasks"
)

// App 备份应用主体，管理整个备份流程
type App struct {
	cfg *config.Config // 应用配置
}

// New 创建备份应用实例
func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

// Run 执行完整的备份流程：检查 -> 备份 -> 打包 -> 清理
func (a *App) Run() error {
	startTime := time.Now()

	timestamp := startTime.Format("20060102_150405")
	slog.Info("🚀 开始备份", "timestamp", timestamp)

	s := scheduler.New(a.cfg)

	// 阶段1: 环境检查和准备
	s.Register(
		&tasks.CheckDataDir{},       // 检查数据目录是否存在
		&tasks.CheckDiskSpace{},     // 检查磁盘空间是否充足
		&tasks.CreateBackupTmpDir{}, // 创建临时备份目录
	)

	// 阶段2: 执行数据备份任务
	s.Register(
		&tasks.DatabaseTask{},                // 备份 SQLite 数据库
		&tasks.RSATask{},                     // 备份 RSA 密钥文件
		&tasks.CopyTask{Path: ".env"},        // 备份环境配置文件
		&tasks.CopyTask{Path: "config.json"}, // 备份应用配置文件
		&tasks.CopyTask{Path: "attachments"}, // 备份附件目录
		&tasks.CopyTask{Path: "sends"},       // 备份发送文件目录
	)

	// 阶段3: 打包压缩和加密
	s.Register(&tasks.ArchiveTask{
		Timestamp: timestamp,
	})

	// 阶段4: 清理过期备份
	s.Register(&tasks.CleanupTask{})

	// 确保临时目录在函数结束时被清理
	defer func() {
		slog.Debug("🧽 清理临时文件", "tmpDir", a.cfg.TmpDir)
		os.RemoveAll(a.cfg.TmpDir)
	}()

	if err := s.Start(); err != nil {
		slog.Error("🚨 备份失败", "error", err)
		return err
	}

	duration := time.Since(startTime)
	slog.Info("✅ 备份完成", "duration", duration)
	return nil
}
