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
	entries, err := os.ReadDir(cfg.TmpDir)
	if err != nil {
		return fmt.Errorf("读取备份目录失败: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("备份目录为空")
	}

	archiveFile := filepath.Join(cfg.BackupDir, fmt.Sprintf("%s_%s.tar.gz", cfg.Filename, c.Timestamp))
	slog.Debug("🔐 创建加密归档", "file", filepath.Base(archiveFile))

	// 创建加密归档
	if err := archive.EncryptedBackup(cfg.TmpDir, cfg.Password, archiveFile); err != nil {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("创建加密归档失败: %w", err)
	}

	// 验证归档完整性 - 解密到单独的验证目录
	verifyDir := filepath.Join(cfg.BackupDir, "verify_tmp")
	defer utils.RemoveIfExists(verifyDir) // 确保验证目录被清理

	if err := utils.EnsureDir(verifyDir); err != nil {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("创建验证目录失败: %w", err)
	}

	if err := archive.DecryptBackup(archiveFile, cfg.Password, verifyDir); err != nil {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("归档验证失败: %w", err)
	}

	slog.Debug("✅ 归档验证成功", "file", filepath.Base(archiveFile))
	return nil
}
