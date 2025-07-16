package engine

import (
	"log/slog"
	"sync"
	"time"

	"github.com/xg4/vaultwarden-backup/internal/config"
)

// Scheduler 任务调度器，支持串行和并行执行任务
type Scheduler struct {
	cfg   *config.Config // 配置信息
	tasks [][]Task       // 任务组列表，每组内的任务并行执行，组间串行执行
}

// New 创建任务调度器实例
func New(cfg *config.Config) *Scheduler {
	return &Scheduler{
		cfg: cfg,
	}
}

// Register 注册任务到调度器
// 同一次调用的任务会并行执行，不同次调用的任务组会串行执行
func (s *Scheduler) Register(t ...Task) {
	s.tasks = append(s.tasks, t)
}

// Start 开始执行所有已注册的任务
func (s *Scheduler) Start() error {
	for _, taskSlice := range s.tasks {
		if len(taskSlice) == 1 {
			// 单个任务直接执行
			task := taskSlice[0]
			if err := handleTask(task, s.cfg); err != nil {
				return err
			}
		} else {
			// 多个任务并行执行
			if err := s.runConcurrentTasks(taskSlice); err != nil {
				return err
			}
		}
	}
	return nil
}

// handleTask 执行单个任务并记录执行时间和结果
func handleTask(t Task, cfg *config.Config) error {
	slog.Debug("开始执行", "task", t.Name())
	start := time.Now()
	err := t.Run(cfg)
	duration := time.Since(start)
	if err != nil {
		slog.Error("执行失败", "task", t.Name(), "error", err)
		return err
	}
	slog.Info("执行成功", "task", t.Name(), "duration", duration)
	return nil
}

// runConcurrentTasks 并行执行多个任务
func (s *Scheduler) runConcurrentTasks(tasks []Task) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))

	// 启动所有任务的 goroutine
	for _, task := range tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()
			if err := handleTask(t, s.cfg); err != nil {
				errChan <- err
				return
			}
		}(task)
	}

	// 等待所有任务完成
	wg.Wait()
	close(errChan)

	// 检查是否有任务执行失败
	for err := range errChan {
		if err != nil {
			return err // 返回第一个遇到的错误
		}
	}

	return nil
}
