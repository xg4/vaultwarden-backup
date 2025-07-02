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
	Timestamp      string
	RetentionDays  int
	Password       string
	MaxConcurrency int
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

	maxConcurrencyStr := getEnv("MAX_CONCURRENCY", "6")
	maxConcurrency, err := strconv.Atoi(maxConcurrencyStr)
	if err != nil {
		return nil, fmt.Errorf("无效的 MAX_CONCURRENCY: %v", err)
	}
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	backupIntervalStr := getEnv("BACKUP_INTERVAL", "6h")
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
		Timestamp:      time.Now().Format("20060102_150405"),
		RetentionDays:  retentionDays,
		Password:       password,
		MaxConcurrency: maxConcurrency,
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
