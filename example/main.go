package main

import (
	"context"
	"fmt"

	worker "github.com/brian-nunez/task-orchestration"
)

type PrintTask struct {
	Message string
}

func (p *PrintTask) Process(ctx context.Context, pc *worker.ProcessContext) error {
	_ = pc.Logger(p.Message)
	return nil
}

type ErrorTask struct {
	Message string
}

func (p *ErrorTask) Process(ctx context.Context, pc *worker.ProcessContext) error {
	_ = pc.Logger(p.Message)
	return fmt.Errorf("error processing task %v %v", p.Message, pc.ProcessId)
}

type PanicTask struct {
	Message string
}

func (p *PanicTask) Process(ctx context.Context, pc *worker.ProcessContext) error {
	_ = pc.Logger(p.Message)

	panic("panic in task")
}

func main() {
	pool := worker.NewWorkerPool(worker.WorkerPoolConfig{
		Concurrency:  10,
		LogPath:      "logs",
		DatabasePath: "./tasks.db",
	})

	err := pool.Start()
	if err != nil {
		panic(err)
	}
	defer pool.Stop()

	for i := 0; i < 200; i++ {
		if i == 160 {
			task := &PanicTask{
				Message: fmt.Sprintf("Task %d", i),
			}

			_, _ = pool.AddTask(task)

			continue
		}
		if i%30 == 0 {
			task := &ErrorTask{
				Message: fmt.Sprintf("Task %d", i),
			}

			_, _ = pool.AddTask(task)

			continue
		}
		task := &PrintTask{
			Message: fmt.Sprintf("Task %d", i),
		}

		_, _ = pool.AddTask(task)
	}

	pool.Wait()

	tasks, err := pool.GetAllTasks()

	for _, t := range *tasks {
		fmt.Printf(
			"Process ID: %v, Status: %v\n",
			t.ProcessID,
			t.Status,
		)
	}

	pool.Wait()

	_, _ = pool.AddTask(&PrintTask{
		Message: "LATER TASK",
	})

	pool.Wait()
}
