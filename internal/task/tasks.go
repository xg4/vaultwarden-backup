package task

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/utils"
)

// Context 包含了执行任务所需的所有依赖
type Context struct {
	Cfg *config.Config
	Log *slog.Logger
}

// Task 定义了一个备份任务单元
type Task struct {
	Name     string
	ActionFn func(ctx *Context) error
}

// Execute 执行单个任务
func (t *Task) Execute(ctx *Context) error {
	return t.ActionFn(ctx)
}

// GetAllTasks 返回所有预定义的备份任务列表
func GetAllTasks() []Task {
	return []Task{
		{Name: "数据库", ActionFn: backupDatabase},
		{Name: "RSA密钥", ActionFn: backupRSAKeys},
		{Name: "配置文件(.env)", ActionFn: backupEnvFile},
		{Name: "配置文件(config.json)", ActionFn: backupConfigJSON},
		{Name: "附件目录(attachments)", ActionFn: backupAttachments},
		{Name: "发送目录(sends)", ActionFn: backupSends},
	}
}

func backupDatabase(ctx *Context) error {
	srcDB := filepath.Join(ctx.Cfg.DataDir, "db.sqlite3")
	destDB := filepath.Join(ctx.Cfg.BackupTmpDir, "db.sqlite3")

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

func backupConfigJSON(ctx *Context) error {
	return copyItem(ctx, "config.json", false)
}

func backupEnvFile(ctx *Context) error {
	return copyItem(ctx, ".env", false)
}

func backupAttachments(ctx *Context) error {
	return copyItem(ctx, "attachments", true)
}

func backupSends(ctx *Context) error {
	return copyItem(ctx, "sends", true)
}

func backupRSAKeys(ctx *Context) error {
	matches, err := filepath.Glob(filepath.Join(ctx.Cfg.DataDir, "rsa_key*"))
	if err != nil {
		return fmt.Errorf("查找RSA密钥失败: %w", err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("RSA密钥不存在")
	}

	for _, file := range matches {
		if err := utils.CopyFile(file, filepath.Join(ctx.Cfg.BackupTmpDir, filepath.Base(file))); err != nil {
			return fmt.Errorf("备份RSA密钥 %s 失败: %w", file, err)
		}
	}
	return nil
}

func copyItem(ctx *Context, name string, isDir bool) error {
	src := filepath.Join(ctx.Cfg.DataDir, name)
	dest := filepath.Join(ctx.Cfg.BackupTmpDir, name)

	if _, err := os.Stat(src); os.IsNotExist(err) {
		ctx.Log.Warn("文件或目录不存在或为空", "name", name)
		return nil
	}

	ctx.Log.Debug(fmt.Sprintf("备份 %s -> %s", src, dest))

	if isDir {
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
