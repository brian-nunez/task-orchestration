package worker

import (
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

func (ctx *ProcessContext) Logger(message any) {
	content := fmt.Sprintf("[WORKER %d]: %v\n", ctx.WorkerId, message)

	_ = ctx.WriteToLogFile(content)
}

func (ctx *ProcessContext) WriteToLogFile(content string) error {
	filepath := filepath.Join(ctx.LogPath, fmt.Sprintf("%s.log", ctx.ProcessId))

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err.Error())
		return err
	}
	defer file.Close()

	file.WriteString(content)

	return nil
}

func (ctx *ProcessContext) SafeProcess(task Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during task execution: %v", r)
			ctx.Logger(err)
		}
	}()

	err = task.Process(ctx)

	return err
}
