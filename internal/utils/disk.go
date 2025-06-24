package utils

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

type DiskSpaceCtx struct {
	SrcDir string // 数据目录
	DstDir string // 备份目录
	Log    *slog.Logger
}

func CheckDiskSpace(ctx *DiskSpaceCtx) error {
	var dataSize int64
	err := filepath.Walk(ctx.SrcDir, func(_ string, info os.FileInfo, err error) error {
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

	var stat unix.Statfs_t
	if err := unix.Statfs(ctx.DstDir, &stat); err != nil {
		return fmt.Errorf("获取文件系统状态失败: %w", err)
	}

	availableSpace := int64(stat.Bavail) * stat.Bsize
	requiredSpace := dataSize * 2

	ctx.Log.Debug("检查磁盘空间", "需要", formatBytes(requiredSpace), "可用", formatBytes(availableSpace))

	if availableSpace < requiredSpace {
		return fmt.Errorf("磁盘空间不足。需要: %s, 可用: %s", formatBytes(requiredSpace), formatBytes(availableSpace))
	}

	return nil
}

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
