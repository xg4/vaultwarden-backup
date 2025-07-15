package engine

import (
	"log/slog"
	"sync"
	"time"

	"github.com/xg4/vaultwarden-backup/internal/config"
)

type Scheduler struct {
	cfg   *config.Config
	tasks [][]Task
}

func New(cfg *config.Config) *Scheduler {
	return &Scheduler{
		cfg: cfg,
	}
}

func (s *Scheduler) Register(t ...Task) {
	s.tasks = append(s.tasks, t)
}

func (s *Scheduler) Start() error {
	for _, taskSlice := range s.tasks {
		if len(taskSlice) == 1 {
			// 单个任务，直接执行
			task := taskSlice[0]
			if err := handleTask(task, s.cfg); err != nil {
				return err // 直接返回错误并退出
			}
		} else {
			// 多个任务，并行执行
			if err := s.runConcurrentTasks(taskSlice); err != nil {
				return err // 直接返回错误并退出
			}
		}
	}
	return nil
}

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

func (s *Scheduler) runConcurrentTasks(tasks []Task) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))

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

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			return err // 返回第一个遇到的错误
		}
	}

	return nil
}
