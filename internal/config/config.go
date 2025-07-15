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
	BackupDir      string
	BackupTmpDir   string
	DataDir        string
	Filename       string
	RetentionDays  int
	Password       string
	BackupInterval time.Duration
}

// Load 从环境变量中加载配置
func Load() (*Config, error) {
	password := os.Getenv("PASSWORD")
	if strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("错误：未设置 PASSWORD 环境变量。请设置备份密码：export PASSWORD='your_password'")
	}

	retentionDaysStr := getEnv("RETENTION_DAYS", "30")
	retentionDays, err := strconv.Atoi(retentionDaysStr)
	if err != nil {
		return nil, fmt.Errorf("无效的 RETENTION_DAYS: %v", err)
	}

	backupIntervalStr := getEnv("BACKUP_INTERVAL", "1h")
	backupInterval, err := time.ParseDuration(backupIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("无效的 BACKUP_INTERVAL: %v", err)
	}

	backupDir := getEnv("BACKUP_DIR", "/backups")
	backupTmpDir := filepath.Join(backupDir, "tmp")

	cfg := &Config{
		BackupDir:      backupDir,
		BackupTmpDir:   backupTmpDir,
		DataDir:        getEnv("DATA_DIR", "/data"),
		Filename:       getEnv("FILENAME", "vault"),
		RetentionDays:  retentionDays,
		Password:       password,
		BackupInterval: backupInterval,
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
