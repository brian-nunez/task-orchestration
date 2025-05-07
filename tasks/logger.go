package tasks

import (
	"fmt"
	"time"

	"github.com/brian-nunez/task-orchestration/worker"
)

type LoggerTask struct {
	Text     string
	LogLevel string
	Delay    time.Duration
}

func (task *LoggerTask) Process(taskContext worker.ProcessContext) error {
	time.Sleep(task.Delay)
	taskContext.Logger(fmt.Sprintf("[%s]: %s\n", task.LogLevel, task.Text))

	return nil
}
