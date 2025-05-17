package tasks

import (
	"fmt"
	"os/exec"

	worker "github.com/brian-nunez/task-orchestration"
)

type M3U8Task struct {
	URL    string
	Output string
}

func (task *M3U8Task) Process(taskContext *worker.ProcessContext) error {
	cmd := exec.Command(
		"m3u8-cli",
		"--url",
		task.URL,
		"--output",
		task.Output,
	)

	taskContext.Logger(fmt.Sprintf("Executing command: %v\n", cmd.String()))

	cmd.Stdin = taskContext.Stdin
	cmd.Stdout = taskContext.Stdout
	cmd.Stderr = taskContext.Stderr

	err := cmd.Run()
	if err != nil {
		taskContext.Logger(fmt.Sprintf("Error executing m3u8-cli: %v", err.Error()))
		return err
	}

	return nil
}
