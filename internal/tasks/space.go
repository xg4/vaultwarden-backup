package tasks

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/xg4/vaultwarden-backup/internal/config"
	"golang.org/x/sys/unix"
)

// CheckDiskSpace 磁盘空间检查任务
type CheckDiskSpace struct{}

func (CheckDiskSpace) Name() string {
	return "检查磁盘空间"
}

// Run 检查备份目录是否有足够的磁盘空间
// 需要的空间 = 数据目录大小 * 2（考虑压缩和临时文件）
func (CheckDiskSpace) Run(cfg *config.Config) error {
	src, dst := cfg.DataDir, cfg.BackupDir

	// 计算数据目录总大小
	var dataSize int64
	err := filepath.Walk(src, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			dataSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("计算数据目录大小时出错: %w", err)
	}

	// 获取备份目录所在文件系统的可用空间
	var stat unix.Statfs_t
	if err := unix.Statfs(dst, &stat); err != nil {
		return fmt.Errorf("获取文件系统状态失败: %w", err)
	}

	availableSpace := int64(stat.Bavail) * int64(stat.Bsize) // 修复类型转换问题
	requiredSpace := dataSize * 2                            // 预留2倍空间用于压缩和临时文件

	slog.Debug("💾 磁盘空间检查", "required", formatBytes(requiredSpace), "available", formatBytes(availableSpace))

	if availableSpace < requiredSpace {
		return fmt.Errorf("磁盘空间不足: 需要 %s, 可用 %s", formatBytes(requiredSpace), formatBytes(availableSpace))
	}

	return nil
}

// formatBytes 将字节数格式化为人类可读的格式（B, KB, MB, GB, TB, PB, EB）
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
