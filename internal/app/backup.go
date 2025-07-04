package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/xg4/vaultwarden-backup/internal/archive"
	"github.com/xg4/vaultwarden-backup/internal/config"
	"github.com/xg4/vaultwarden-backup/internal/tasks"
	"github.com/xg4/vaultwarden-backup/internal/utils"
)

// App 封装了备份应用的状态和依赖
type App struct {
	cfg       *config.Config
	Timestamp string
}

// New 创建一个新的 App 实例
func New(cfg *config.Config, startTime time.Time) *App {
	return &App{cfg: cfg, Timestamp: startTime.Format("20060102_150405")}
}

// Run 执行完整的备份和清理流程
func (a *App) Run() error {
	slog.Info("开始备份流程", "timestamp", a.Timestamp)

	// 1. 运行前检查
	if err := utils.ValidateDirectories(a.cfg); err != nil {
		return err
	}

	// 检查磁盘空间
	if err := utils.CheckDiskSpace(a.cfg.DataDir, a.cfg.BackupDir); err != nil {
		return err
	}

	// 2. 初始化备份环境
	if err := a.initBackupDir(); err != nil {
		return err
	}

	// 使用 defer 确保临时目录总能被清理
	defer os.RemoveAll(a.cfg.BackupTmpDir)

	// 3. 执行所有备份任务
	if err := a.runTasks(); err != nil {
		return err
	}

	// 4. 打包压缩和加密
	if err := a.createArchive(); err != nil {
		return fmt.Errorf("创建归档文件失败: %w", err)
	}

	// 5. 清理旧备份
	if err := a.cleanupOldBackups(); err != nil {
		// 清理旧备份的失败不应导致整个备份失败，只记录警告
		slog.Warn("清理旧备份失败", "error", err)
	}

	return nil
}

func (a *App) initBackupDir() error {
	// 安全地清理并创建备份目录
	if err := utils.RemoveIfExists(a.cfg.BackupTmpDir); err != nil {
		return fmt.Errorf("无法清理现有备份目录: %w", err)
	}
	if err := utils.EnsureDir(a.cfg.BackupTmpDir); err != nil {
		return fmt.Errorf("无法创建备份目录: %w", err)
	}
	return nil
}

type TaskResult struct {
	TaskName string
	Error    error
	Duration time.Duration
}

type taskJob struct {
	task  tasks.Task
	index int
}

func (a *App) runTasks() error {
	taskCtx := &tasks.Context{Cfg: a.cfg}
	tasks := tasks.GetAllTasks()

	// 创建带取消功能的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobChan := make(chan taskJob, len(tasks))
	resultChan := make(chan TaskResult, len(tasks))
	// 用于收集错误的channel
	errorChan := make(chan error, 1)
	// 用于等待所有goroutine完成
	var wg sync.WaitGroup

	slog.Info("执行备份任务", "concurrency", a.cfg.MaxConcurrency, "total", len(tasks))

	maxWorkers := min(a.cfg.MaxConcurrency, len(tasks))

	for i := range maxWorkers {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case job, ok := <-jobChan:
					if !ok {
						return
					}

					slog.Debug("执行任务", "worker", workerID, "task", job.task.Name)
					start := time.Now()
					err := job.task.Execute(taskCtx)
					duration := time.Since(start)

					result := TaskResult{
						TaskName: job.task.Name,
						Error:    err,
						Duration: duration,
					}

					if err != nil {
						// 发生错误时立即通知主协程
						select {
						case errorChan <- fmt.Errorf("任务 [%s] 失败: %w", job.task.Name, err):
						default: // 如果已经有错误了，忽略后续错误
						}
						// 仍然发送结果用于统计
						select {
						case resultChan <- result:
						case <-ctx.Done():
						}
						return // 工作协程退出
					}

					select {
					case resultChan <- result:
					case <-ctx.Done():
						return
					}

				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	// 分发任务协程
	go func() {
		defer close(jobChan)
		for i, t := range tasks {
			select {
			case jobChan <- taskJob{task: t, index: i}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// 等待工作协程完成的协程
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 主控制循环 - 监听错误和结果
	var completedTasks []TaskResult
	totalTasks := len(tasks)

	for {
		select {
		case err := <-errorChan:
			// 有任务失败，立即取消所有任务
			slog.Error("任务执行失败", "error", err, "completed", len(completedTasks), "total", totalTasks)
			cancel() // 取消所有工作协程

			// 等待一小段时间让正在执行的任务有机会完成清理
			time.Sleep(100 * time.Millisecond)

			return err

		case result, ok := <-resultChan:
			if !ok {
				// 所有任务都成功完成
				slog.Info("所有备份任务完成", "completed", len(completedTasks), "total", totalTasks)
				return nil
			}

			completedTasks = append(completedTasks, result)

			if result.Error == nil {
				slog.Info("任务完成", "task", result.TaskName, "duration", result.Duration.String(), "progress", fmt.Sprintf("%d/%d", len(completedTasks), totalTasks))
			}
			// 错误情况已经在 errorChan 中处理了
		}
	}
}

func (a *App) createArchive() error {
	entries, err := os.ReadDir(a.cfg.BackupTmpDir)
	if err != nil {
		return fmt.Errorf("读取备份目录失败: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("备份目录为空")
	}

	archiveFile := filepath.Join(a.cfg.BackupDir, fmt.Sprintf("%s_%s.tar.gz", a.cfg.Filename, a.Timestamp))
	slog.Info("创建加密压缩包", "file", archiveFile)

	if err := archive.EncryptedBackup(a.cfg.BackupTmpDir, a.cfg.Password, archiveFile); err != nil {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("创建加密归档失败: %w", err)
	}

	if err := archive.DecryptBackup(archiveFile, a.cfg.Password, a.cfg.BackupTmpDir); err != nil {
		utils.RemoveIfExists(archiveFile)
		return fmt.Errorf("解密归档失败: %w", err)
	}

	return nil
}

func (a *App) cleanupOldBackups() error {
	if a.cfg.RetentionDays <= 0 {
		return nil
	}

	globPattern := filepath.Join(a.cfg.BackupDir, fmt.Sprintf("%s_*.tar.gz", a.cfg.Filename))

	files, err := filepath.Glob(globPattern)
	if err != nil {
		return fmt.Errorf("查找旧备份失败: %w", err)
	}
	slog.Info("找到备份文件", "count", len(files), "pattern", globPattern)

	cutoffTime := time.Now().AddDate(0, 0, -a.cfg.RetentionDays)
	count := 0
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			slog.Warn("获取文件信息失败", "file", file, "error", err)
			continue
		}
		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(file); err != nil {
				slog.Warn("删除旧备份失败", "file", file, "error", err)
			} else {
				count++
			}
		}
	}

	if count > 0 {
		slog.Info("清理旧备份完成", "count", count, "retention_days", a.cfg.RetentionDays)
	}

	return nil
}
