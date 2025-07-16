package tasks

import (
	"fmt"
	"path/filepath"

	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/utils"
)

// RSATask RSA 密钥文件备份任务
type RSATask struct{}

func (RSATask) Name() string { return "备份RSA密钥" }

// Run 备份所有 RSA 密钥相关文件
// 包括 rsa_key*, rsa_key.pem, rsa_key.pub.pem 等文件
func (RSATask) Run(cfg *config.Config) error {
	// 查找所有 RSA 密钥文件
	matches, err := filepath.Glob(filepath.Join(cfg.DataDir, "rsa_key*"))
	if err != nil {
		return fmt.Errorf("查找RSA密钥失败: %w", err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("RSA密钥不存在")
	}

	// 逐个复制密钥文件
	for _, file := range matches {
		destFile := filepath.Join(cfg.BackupTmpDir, filepath.Base(file))
		if err := utils.CopyFile(file, destFile); err != nil {
			return fmt.Errorf("备份RSA密钥 %s 失败: %w", file, err)
		}
	}
	return nil
}
