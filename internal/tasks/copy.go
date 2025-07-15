package tasks

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/utils"
)

type CopyTask struct {
	Path string
}

func (c *CopyTask) Name() string { return "备份" + c.Path }

func (c *CopyTask) Run(cfg *config.Config) error {
	return copyItem(cfg, c.Path)
}

func copyItem(cfg *config.Config, name string) error {
	src := filepath.Join(cfg.DataDir, name)
	dest := filepath.Join(cfg.BackupTmpDir, name)

	fileInfo, err := os.Stat(src)
	if os.IsNotExist(err) {
		slog.Debug("文件或目录不存在", "name", name)
		return nil
	}

	slog.Debug(fmt.Sprintf("备份 %s -> %s", src, dest))

	if fileInfo.IsDir() {
		if err := utils.CopyDir(src, dest); err != nil {
			return fmt.Errorf("%s 备份失败: %w", name, err)
		}
	} else {
		if err := utils.CopyFile(src, dest); err != nil {
			return fmt.Errorf("%s 备份失败: %w", name, err)
		}
	}
	return nil
}
