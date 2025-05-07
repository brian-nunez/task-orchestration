package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

type ProcessContext struct {
	WorkerId  int
	ProcessId string
	Params    any
	Stdin     *os.File
	Stderr    *os.File
	Stdout    *os.File
}

func (ctx *ProcessContext) Logger(message any) {
	content := fmt.Sprintf("[WORKER %d]: %v\n", ctx.WorkerId, message)
	// fmt.Printf(content)

	_ = ctx.WriteToLogFile(content)
}

func (ctx *ProcessContext) WriteToLogFile(content string) error {
	filepath := filepath.Join("logs", fmt.Sprintf("%s.log", ctx.ProcessId))

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err.Error())
		return err
	}
	defer file.Close()

	file.WriteString(content)

	return nil
}

type Task interface {
	Process(taskContext *ProcessContext) error
}

type WorkerPool struct {
	Concurreny int
	tasksChan  chan Task
	wg         sync.WaitGroup
	doneChan   chan struct{}
}

func (wp *WorkerPool) Start() {
	wp.tasksChan = make(chan Task)
	wp.doneChan = make(chan struct{})

	for i := 0; i < wp.Concurreny; i++ {
		go wp.worker(i)
	}
}

func (wp *WorkerPool) AddTask(task Task) {
	wp.wg.Add(1)
	wp.tasksChan <- task
}

func (wp *WorkerPool) Stop() {
	close(wp.tasksChan)
	wp.wg.Wait()
	close(wp.doneChan)
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

func (wp *WorkerPool) worker(workerId int) {
	for task := range wp.tasksChan {
		fmt.Printf("Worker started %d\n", workerId)
		ctx := &ProcessContext{
			WorkerId:  workerId,
			ProcessId: uuid.NewString(),
		}

		filepath := filepath.Join("logs", fmt.Sprintf("%s.log", ctx.ProcessId))

		file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			ctx.Logger(fmt.Sprintf("Error opening log file: %v", err.Error()))
			wp.wg.Done()
			continue
		}

		ctx.Stdout = file
		ctx.Stderr = file
		ctx.Stdin = file

		ctx.Logger("Starting task\n")
		err = task.Process(ctx)
		if err != nil {
			fmt.Printf("[%v]: Error processing task: %v\n", ctx.ProcessId, err.Error())
		} else {
			ctx.Logger("Task completed successfully\n")
		}
		file.Close()
		wp.wg.Done()
		fmt.Printf("Worker %d finished\n", workerId)
	}

}

func (wp *WorkerPool) Run() {
	wp.tasksChan = make(chan Task)

	for i := 0; i < wp.Concurreny; i++ {
		go wp.worker(i)
	}

	// wp.wg.Add(len(wp.Tasks))
	// for _, task := range wp.Tasks {
	// 	wp.tasksChan <- task
	// }
	close(wp.tasksChan)

	wp.wg.Wait()
}
