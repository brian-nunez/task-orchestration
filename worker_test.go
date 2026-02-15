package worker_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	worker "github.com/brian-nunez/task-orchestration"
)

const (
	dbPath = "test.db"
)

func setup(t *testing.T) {
	_ = os.Remove(dbPath)
}

func teardown(t *testing.T) {
	_ = os.Remove(dbPath)
}

type TestTask struct {
	Delay    time.Duration
	Text     string
	Error    bool
	Panic    bool
	Canceled bool
}

func (t *TestTask) Process(ctx context.Context, taskContext *worker.ProcessContext) error {
	select {
	case <-ctx.Done():
		t.Canceled = true
		return ctx.Err()
	case <-time.After(t.Delay):
		if t.Panic {
			panic("test panic")
		}
		if t.Error {
			return fmt.Errorf("test error")
		}
		_ = taskContext.Logger(t.Text)
		return nil
	}
}

func TestWorkerPool(t *testing.T) {
	setup(t)
	defer teardown(t)

	pool := worker.NewWorkerPool(worker.WorkerPoolConfig{
		Concurrency: 2,
		DatabasePath: dbPath,
	})

	err := pool.Start()
	if err != nil {
		t.Fatalf("Error starting worker pool: %v", err)
	}

	// Add a task that will succeed
	_, err = pool.AddTask(&TestTask{Delay: 1 * time.Second, Text: "hello"})
	if err != nil {
		t.Fatalf("Error adding task: %v", err)
	}

	// Add a task that will fail
	_, err = pool.AddTask(&TestTask{Delay: 1 * time.Second, Error: true})
	if err != nil {
		t.Fatalf("Error adding task: %v", err)
	}

	// Add a task that will panic
	_, err = pool.AddTask(&TestTask{Delay: 1 * time.Second, Panic: true})
	if err != nil {
		t.Fatalf("Error adding task: %v", err)
	}

	pool.Wait()

	// Check the status of the tasks
	completedTasks, err := pool.GetCompletedTasks()
	if err != nil {
		t.Fatalf("Error getting completed tasks: %v", err)
	}
	if len(*completedTasks) != 1 {
		t.Errorf("Expected 1 completed task, got %d", len(*completedTasks))
	}

	failedTasks, err := pool.GetFailedTasks()
	if err != nil {
		t.Fatalf("Error getting failed tasks: %v", err)
	}
	if len(*failedTasks) != 2 {
		t.Errorf("Expected 2 failed tasks, got %d", len(*failedTasks))
	}

	pool.Stop()
}

func TestWorkerPool_Cancel(t *testing.T) {
	setup(t)
	defer teardown(t)

	pool := worker.NewWorkerPool(worker.WorkerPoolConfig{
		Concurrency: 2,
		DatabasePath: dbPath,
	})

	err := pool.Start()
	if err != nil {
		t.Fatalf("Error starting worker pool: %v", err)
	}

	task := &TestTask{Delay: 5 * time.Second}
	_, err = pool.AddTask(task)
	if err != nil {
		t.Fatalf("Error adding task: %v", err)
	}

	go func() {
		time.Sleep(1 * time.Second)
		pool.Stop()
	}()

	pool.Wait()

	if !task.Canceled {
		t.Error("Expected task to be canceled")
	}
}
