package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config 保存了应用的所有配置
type Config struct {
	BackupDir         string
	TmpDir            string
	DataDir           string
	BackupName        string
	PruneBackupsDays  int
	PruneBackupsCount int
	Password          string
	BackupInterval    time.Duration
}

// Load 从环境变量中加载配置
func Load() (*Config, error) {
	password := os.Getenv("PASSWORD")
	if strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("错误：未设置 PASSWORD 环境变量。请设置备份密码：export PASSWORD='your_password'")
	}

	pruneBackupsDaysStr := getEnv("PRUNE_BACKUPS_DAYS", "30")
	pruneBackupsDays, err := strconv.Atoi(pruneBackupsDaysStr)
	if err != nil {
		return nil, fmt.Errorf("无效的 PRUNE_BACKUPS_DAYS: %v", err)
	}
	if pruneBackupsDays < 0 {
		pruneBackupsDays = 0
	}

	pruneBackupsCountStr := getEnv("PRUNE_BACKUPS_COUNT", "0")
	pruneBackupsCount, err := strconv.Atoi(pruneBackupsCountStr)
	if err != nil {
		return nil, fmt.Errorf("无效的 PRUNE_BACKUPS_COUNT: %v", err)
	}
	if pruneBackupsCount < 0 {
		pruneBackupsCount = 0
	}

	backupIntervalStr := getEnv("BACKUP_INTERVAL", "6h")
	backupInterval, err := time.ParseDuration(backupIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("无效的 BACKUP_INTERVAL: %v", err)
	}
	if backupInterval < time.Minute {
		backupInterval = time.Minute
	}

	backupDir := getEnv("BACKUP_DIR", "/backups")
	dataDir := getEnv("DATA_DIR", "/data")
	tmpDir := filepath.Join(backupDir, "/.backup_tmp")

	cfg := &Config{
		BackupDir:         backupDir,
		DataDir:           dataDir,
		TmpDir:            tmpDir,
		BackupName:        getEnv("BACKUP_NAME", "vault"),
		PruneBackupsDays:  pruneBackupsDays,
		PruneBackupsCount: pruneBackupsCount,
		Password:          password,
		BackupInterval:    backupInterval,
	}

	return cfg, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
