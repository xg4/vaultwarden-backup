package task

import (
	"log/slog"

	"github.com/xg4/vaultwarden-backup/internal/config"
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
