package main

import (
	"fmt"

	worker "github.com/brian-nunez/task-orchestration"
)

type PrintTask struct {
	Message string
}

func (p *PrintTask) Process(ctx *worker.ProcessContext) error {
	ctx.Logger(p.Message)
	return nil
}

func main() {
	pool := &worker.WorkerPool{
		Concurreny: 10,
		LogPath:    "logs",
	}

	pool.Start()
	defer pool.Stop()

	for i := 0; i < 200; i++ {
		task := &PrintTask{
			Message: fmt.Sprintf("Task %d", i),
		}
		pool.AddTask(task)
	}

	fmt.Println("Finished for loop")

	pool.Wait()

	pool.AddTask(&PrintTask{
		Message: "LATER TASK",
	})

	pool.Wait()
}
