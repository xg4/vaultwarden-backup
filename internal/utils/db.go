package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// BackupSQLite 使用sqlite3命令备份数据库
func BackupSQLite(srcPath, destPath string) error {
	cmd := exec.Command("sqlite3", srcPath, fmt.Sprintf(".backup '%s'", destPath))
	return cmd.Run()
}

// CheckSQLiteIntegrity 检查SQLite数据库完整性
func CheckSQLiteIntegrity(dbPath string) error {
	cmd := exec.Command("sqlite3", dbPath, "PRAGMA integrity_check;")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	result := strings.TrimSpace(string(output))
	if result != "ok" {
		return fmt.Errorf("数据完整性检查发现问题:\n%s", result)
	}
	return nil
}