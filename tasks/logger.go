package tasks

import (
	"context"
	"fmt"
	"time"

	worker "github.com/brian-nunez/task-orchestration"
)

type LoggerTask struct {
	Text     string
	LogLevel string
	Delay    time.Duration
}

func (task *LoggerTask) Process(ctx context.Context, taskContext *worker.ProcessContext) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(task.Delay):
		_ = taskContext.Logger(fmt.Sprintf("[%s]: %s\n", task.LogLevel, task.Text))
		return nil
	}
}
