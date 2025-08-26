package tasks

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/xg4/vaultwarden-backup/internal/config"
)

type CleanupTask struct{}

func (c *CleanupTask) Name() string { return "清理" }

func (c *CleanupTask) Run(cfg *config.Config) error {
	if cfg.PruneBackupsDays <= 0 && cfg.PruneBackupsCount <= 0 {
		return nil
	}

	if cfg.PruneBackupsDays > 0 && cfg.PruneBackupsCount > 0 {
		slog.Warn("⚠️ PRUNE_BACKUPS_DAYS and PRUNE_BACKUPS_COUNT are both set. PRUNE_BACKUPS_COUNT will be used.")
	}

	globPattern := filepath.Join(cfg.BackupDir, fmt.Sprintf("%s_*.tar.gz", cfg.BackupName))
	files, err := filepath.Glob(globPattern)
	if err != nil {
		return fmt.Errorf("查找旧备份失败: %w", err)
	}
	slog.Debug("🔍 扫描备份文件", "found", len(files))

	if cfg.PruneBackupsCount > 0 {
		if len(files) <= cfg.PruneBackupsCount {
			return nil
		}

		// Sort files by modification time, oldest first
		sort.Slice(files, func(i, j int) bool {
			infoI, errI := os.Stat(files[i])
			infoJ, errJ := os.Stat(files[j])
			if errI != nil || errJ != nil {
				return false
			}
			return infoI.ModTime().Before(infoJ.ModTime())
		})

		filesToDelete := files[:len(files)-cfg.PruneBackupsCount]
		count := 0
		for _, file := range filesToDelete {
			if err := os.Remove(file); err != nil {
				slog.Warn("⚠️ 删除失败", "file", filepath.Base(file), "error", err)
			} else {
				count++
			}
		}

		if count > 0 {
			slog.Info("🧹 清理过期备份", "deleted", count, "keep_count", cfg.PruneBackupsCount)
		}

		return nil
	}

	if cfg.PruneBackupsDays > 0 {
		cutoffTime := time.Now().AddDate(0, 0, -cfg.PruneBackupsDays)
		count := 0
		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				slog.Warn("⚠️ 无法读取文件信息", "file", filepath.Base(file), "error", err)
				continue
			}
			if info.ModTime().Before(cutoffTime) {
				if err := os.Remove(file); err != nil {
					slog.Warn("⚠️ 删除失败", "file", filepath.Base(file), "error", err)
				} else {
					count++
				}
			}
		}

		if count > 0 {
			slog.Info("🧹 清理过期备份", "deleted", count, "prune_backups_days", cfg.PruneBackupsDays)
		}
	}

	return nil
}
