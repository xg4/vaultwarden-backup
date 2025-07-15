package tasks

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/utils"
)

type DatabaseTask struct{}

func (DatabaseTask) Name() string { return "备份数据库" }

func (DatabaseTask) Run(cfg *config.Config) error {
	srcDB := filepath.Join(cfg.DataDir, "db.sqlite3")
	destDB := filepath.Join(cfg.BackupTmpDir, "db.sqlite3")

	if _, err := os.Stat(srcDB); os.IsNotExist(err) {
		return fmt.Errorf("数据库文件 %s 不存在", srcDB)
	}

	if err := utils.BackupSQLite(srcDB, destDB); err != nil {
		return err
	}

	if err := utils.CheckSQLiteIntegrity(destDB); err != nil {
		return err
	}

	return nil
}
