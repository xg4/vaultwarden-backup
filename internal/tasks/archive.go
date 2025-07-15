package tasks

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/xg4/vaultwarden-backup/internal/archive"
	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/utils"
)

type ArchiveTask struct {
	Timestamp string
}

func (c *ArchiveTask) Name() string { return "打包/压缩/加密" }

func (c *ArchiveTask) Run(cfg *config.Config) error {
	entries, err := os.ReadDir(cfg.BackupTmpDir)
	if err != nil {
		return fmt.Errorf("读取备份目录失败: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("备份目录为空")
	}

	archiveFile := filepath.Join(cfg.BackupDir, fmt.Sprintf("%s_%s.tar.gz", cfg.Filename, c.Timestamp))
	slog.Info("创建加密压缩包", "file", archiveFile)

	if err := archive.EncryptedBackup(cfg.BackupTmpDir, cfg.Password, archiveFile); err != nil {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("创建加密归档失败: %w", err)
	}

	if err := archive.DecryptBackup(archiveFile, cfg.Password, cfg.BackupTmpDir); err != nil {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("解密归档失败: %w", err)
	}

	return nil
}
