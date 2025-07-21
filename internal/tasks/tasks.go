package tasks

import (
	"fmt"
	"os"

	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/utils"
)

type CreateBackupTmpDir struct{}

func (CreateBackupTmpDir) Name() string {
	return "创建临时目录"
}

func (CreateBackupTmpDir) Run(cfg *config.Config) error {
	// 安全地清理并创建备份目录
	if err := utils.RemoveIfExists(cfg.BackupTmpDir); err != nil {
		return fmt.Errorf("无法清理临时备份目录: %s, 错误: %v", cfg.BackupTmpDir, err)
	}
	if err := utils.EnsureDir(cfg.BackupTmpDir); err != nil {
		return fmt.Errorf("无法创建临时备份目录: %s, 错误: %v", cfg.BackupTmpDir, err)
	}
	return nil
}

type CheckDataDir struct{}

func (CheckDataDir) Name() string {
	return "检查数据目录"
}

func (CheckDataDir) Run(cfg *config.Config) error {
	info, err := os.Stat(cfg.DataDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("数据目录不存在: %s", cfg.DataDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("数据路径不是一个目录: %s", cfg.DataDir)
	}

	// 检查数据目录是否可读
	if file, err := os.Open(cfg.DataDir); err != nil {
		return fmt.Errorf("无法访问数据目录: %s, 错误: %v", cfg.DataDir, err)
	} else {
		file.Close()
	}

	return nil
}
