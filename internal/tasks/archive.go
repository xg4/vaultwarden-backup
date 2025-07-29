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

	// 验证归档完整性
	slog.Debug("🔎 开始验证归档完整性")

	// 计算原始备份目录的哈希值
	slog.Debug("🧮 计算原始目录哈希", "dir", cfg.TmpDir)
	sourceHash, err := utils.HashDir(cfg.TmpDir)
	if err != nil {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("计算原始目录哈希失败: %w", err)
	}
	slog.Debug("✨ 原始目录哈希", "hash", sourceHash)

	// 解密到单独的验证目录
	verifyDir := filepath.Join(cfg.BackupDir, "/.verify_tmp")
	defer utils.RemoveIfExists(verifyDir) // 确保验证目录被清理

	if err := utils.EnsureDir(verifyDir); err != nil {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("创建验证目录失败: %w", err)
	}

	if err := archive.DecryptBackup(archiveFile, cfg.Password, verifyDir); err != nil {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("解密归档失败: %w", err)
	}

	// 计算验证目录的哈希值
	slog.Debug("🧮 计算验证目录哈希", "dir", verifyDir)
	verifyHash, err := utils.HashDir(verifyDir)
	if err != nil {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("计算验证目录哈希失败: %w", err)
	}
	slog.Debug("✨ 验证目录哈希", "hash", verifyHash)

	// 比较哈希值
	if sourceHash != verifyHash {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("归档完整性验证失败: 哈希值不匹配")
	}

	slog.Debug("✅ 归档验证成功", "file", filepath.Base(archiveFile))
	return nil
}
