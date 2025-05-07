package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/brian-nunez/task-orchestration/tasks"
	"github.com/brian-nunez/task-orchestration/worker"
)

func main() {
	pool := worker.WorkerPool{
		Concurreny: 5,
	}

	pool.Start()

	for i := 0; i < 20; i++ {
		timeToWait := time.Duration(rand.Intn(5) + 1)
		task := &tasks.LoggerTask{
			Text:     fmt.Sprintf("Task %d -- Time to Wait %v", i, timeToWait),
			Delay:    time.Second * timeToWait,
			LogLevel: "INFO",
		}
		pool.AddTask(task)
	}

	time.Sleep(time.Second * 5)
	pool.AddTask(&tasks.LoggerTask{
		Text:     "LATER TASK",
		Delay:    time.Second * 20,
		LogLevel: "INFO",
	})

	pool.Wait()
	pool.Stop()

	// pool.Run()
}
