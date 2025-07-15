package tasks

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/xg4/vaultwarden-backup/internal/config"
)

type CleanupTask struct{}

func (c *CleanupTask) Name() string { return "清理过期备份" }

func (c *CleanupTask) Run(cfg *config.Config) error {
	if cfg.RetentionDays <= 0 {
		return nil
	}

	globPattern := filepath.Join(cfg.BackupDir, fmt.Sprintf("%s_*.tar.gz", cfg.Filename))

	files, err := filepath.Glob(globPattern)
	if err != nil {
		return fmt.Errorf("查找旧备份失败: %w", err)
	}
	slog.Info("找到备份文件", "count", len(files), "pattern", globPattern)

	cutoffTime := time.Now().AddDate(0, 0, -cfg.RetentionDays)
	count := 0
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			slog.Warn("获取文件信息失败", "file", file, "error", err)
			continue
		}
		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(file); err != nil {
				slog.Warn("删除旧备份失败", "file", file, "error", err)
			} else {
				count++
			}
		}
	}

	if count > 0 {
		slog.Info("清理旧备份完成", "count", count, "retention_days", cfg.RetentionDays)
	}

	return nil
}
