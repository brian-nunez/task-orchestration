package tasks

import (
	"context"
	"fmt"
	"os/exec"

	worker "github.com/brian-nunez/task-orchestration"
)

type M3U8Task struct {
	URL    string
	Output string
}

func (task *M3U8Task) Process(ctx context.Context, taskContext *worker.ProcessContext) error {
	cmd := exec.CommandContext(
		ctx,
		"m3u8-cli",
		"--url",
		task.URL,
		"--output",
		task.Output,
	)

	_ = taskContext.Logger(fmt.Sprintf("Executing command: %v\n", cmd.String()))

	cmd.Stdin = taskContext.Stdin
	cmd.Stdout = taskContext.Stdout
	cmd.Stderr = taskContext.Stderr

	err := cmd.Run()
	if err != nil {
		_ = taskContext.Logger(fmt.Sprintf("Error executing m3u8-cli: %v", err.Error()))
		return err
	}

	return nil
}
