package worker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type ProcessContext struct {
	WorkerId  int
	ProcessId string
	Params    any
	Stdin     *os.File
	Stderr    *os.File
	Stdout    *os.File
	LogPath   string
}

func (ctx *ProcessContext) Logger(message any) error {
	content := fmt.Sprintf("[WORKER %d]: %v\n", ctx.WorkerId, message)

	return ctx.WriteToLogFile(content)
}

func (ctx *ProcessContext) WriteToLogFile(content string) error {
	filepath := filepath.Join(ctx.LogPath, fmt.Sprintf("%s.log", ctx.ProcessId))

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err.Error())
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func (pc *ProcessContext) SafeProcess(ctx context.Context, task Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during task execution: %v", r)
			_ = pc.Logger(err)
		}
	}()

	err = task.Process(ctx, pc)

	return err
}