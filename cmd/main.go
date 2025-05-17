package main

import (
	"fmt"
	"math/rand"
	"time"

	worker "github.com/brian-nunez/task-orchestration"
	"github.com/brian-nunez/task-orchestration/tasks"
)

func main() {
	pool := worker.WorkerPool{
		Concurreny: 10,
		LogPath:    "logs",
	}

	pool.Start()

	for i := 0; i < 20; i++ {
		timeToWait := time.Duration(rand.Intn(5)+1) * time.Second
		task := &tasks.LoggerTask{
			Text:     fmt.Sprintf("Task %d -- Time to Wait %v", i, timeToWait),
			Delay:    timeToWait,
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
}
