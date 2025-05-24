package main

import (
	"fmt"

	worker "github.com/brian-nunez/task-orchestration"
	// "github.com/brian-nunez/task-orchestration/state"
)

type PrintTask struct {
	Message string
}

func (p *PrintTask) Process(ctx *worker.ProcessContext) error {
	ctx.Logger(p.Message)
	return nil
}

type ErrorTask struct {
	Message string
}

func (p *ErrorTask) Process(ctx *worker.ProcessContext) error {
	ctx.Logger(p.Message)
	return fmt.Errorf("error processing task %v %v", p.Message, ctx.ProcessId)
}

type PanicTask struct {
	Message string
}

func (p *PanicTask) Process(ctx *worker.ProcessContext) error {
	ctx.Logger(p.Message)

	panic("panic in task")

	return nil
}

func main() {
	pool := &worker.WorkerPool{
		Concurreny:   10,
		LogPath:      "logs",
		DatabasePath: "./tasks.db",
	}

	pool.Start()
	defer pool.Stop()

	for i := 0; i < 200; i++ {
		if i == 160 {
			task := &PanicTask{
				Message: fmt.Sprintf("Task %d", i),
			}

			pool.AddTask(task)

			continue
		}
		if i%30 == 0 {
			task := &ErrorTask{
				Message: fmt.Sprintf("Task %d", i),
			}

			pool.AddTask(task)

			continue
		}
		task := &PrintTask{
			Message: fmt.Sprintf("Task %d", i),
		}

		pool.AddTask(task)
	}

	tasks, err := pool.GetCompletedTasks()

	fmt.Println("Finished for loop", tasks, err)

	pool.Wait()

	pool.AddTask(&PrintTask{
		Message: "LATER TASK",
	})

	pool.Wait()
}
