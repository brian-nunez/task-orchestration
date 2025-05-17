package tasks

import (
	"fmt"
	"time"

	worker "github.com/brian-nunez/task-orchestration"
)

type LoggerTask struct {
	Text     string
	LogLevel string
	Delay    time.Duration
}

func (task *LoggerTask) Process(taskContext *worker.ProcessContext) error {
	fmt.Printf("Delaying for %v\n", task.Delay)
	time.Sleep(task.Delay)
	taskContext.Logger(fmt.Sprintf("[%s]: %s\n", task.LogLevel, task.Text))

	return nil
}
