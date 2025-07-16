package tasks

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/utils"
)

// DatabaseTask SQLite 数据库备份任务
type DatabaseTask struct{}

func (DatabaseTask) Name() string { return "备份数据库" }

// Run 备份 SQLite 数据库文件并验证完整性
func (DatabaseTask) Run(cfg *config.Config) error {
	srcDB := filepath.Join(cfg.DataDir, "db.sqlite3")
	destDB := filepath.Join(cfg.BackupTmpDir, "db.sqlite3")

	// 检查源数据库文件是否存在
	if _, err := os.Stat(srcDB); os.IsNotExist(err) {
		return fmt.Errorf("数据库文件 %s 不存在", srcDB)
	}

	// 使用 sqlite3 命令进行数据库备份
	if err := utils.BackupSQLite(srcDB, destDB); err != nil {
		return err
	}

	// 验证备份文件的完整性
	if err := utils.CheckSQLiteIntegrity(destDB); err != nil {
		return err
	}

	return nil
}
