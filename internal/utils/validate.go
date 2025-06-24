package utils

import (
	"fmt"
	"os"

	"github.com/xg4/vaultwarden-backup/internal/config"
)

func ValidateDirectories(cfg *config.Config) error {
	info, err := os.Stat(cfg.DataDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("数据目录不存在: %s", cfg.DataDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("数据路径不是一个目录: %s", cfg.DataDir)
	}

	return nil
}
