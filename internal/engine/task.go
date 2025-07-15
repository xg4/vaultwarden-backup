package engine

import (
	"github.com/xg4/vaultwarden-backup/internal/config"
)

// Task 定义了一个备份任务单元
type Task interface {
	Name() string
	Run(ctx *config.Config) error
}
