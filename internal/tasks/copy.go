package tasks

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/utils"
)

// CopyTask 文件/目录复制任务
type CopyTask struct {
	Path string // 要备份的文件或目录相对路径
}

func (c *CopyTask) Name() string { return "备份" + c.Path }

// Run 执行文件或目录的复制备份
func (c *CopyTask) Run(cfg *config.Config) error {
	return copyItem(cfg, c.Path)
}

// copyItem 复制指定的文件或目录到备份临时目录
func copyItem(cfg *config.Config, name string) error {
	src := filepath.Join(cfg.DataDir, name)
	dest := filepath.Join(cfg.BackupTmpDir, name)

	// 检查源文件/目录是否存在
	fileInfo, err := os.Stat(src)
	if os.IsNotExist(err) {
		slog.Debug("跳过不存在的文件", "path", name)
		return nil // 文件不存在时不报错，只是跳过
	}

	// 根据文件类型选择复制方式
	if fileInfo.IsDir() {
		// 复制整个目录
		if err := utils.CopyDir(src, dest); err != nil {
			return fmt.Errorf("%s 备份失败: %w", name, err)
		}
	} else {
		// 复制单个文件
		if err := utils.CopyFile(src, dest); err != nil {
			return fmt.Errorf("%s 备份失败: %w", name, err)
		}
	}
	return nil
}
