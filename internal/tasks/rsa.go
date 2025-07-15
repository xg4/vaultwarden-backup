package tasks

import (
	"fmt"
	"path/filepath"

	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/utils"
)

type RSATask struct{}

func (RSATask) Name() string { return "备份RSA密钥" }

func (RSATask) Run(cfg *config.Config) error {
	matches, err := filepath.Glob(filepath.Join(cfg.DataDir, "rsa_key*"))
	if err != nil {
		return fmt.Errorf("查找RSA密钥失败: %w", err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("RSA密钥不存在")
	}

	for _, file := range matches {
		if err := utils.CopyFile(file, filepath.Join(cfg.BackupTmpDir, filepath.Base(file))); err != nil {
			return fmt.Errorf("备份RSA密钥 %s 失败: %w", file, err)
		}
	}
	return nil
}
