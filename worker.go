package worker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/brian-nunez/task-orchestration/state"
	"github.com/brian-nunez/task-orchestration/storage"
	"github.com/google/uuid"
)

const (
	ErrNoRows = "sql: no rows in result set"
)

type Task interface {
	Process(ctx context.Context, taskContext *ProcessContext) error
}

type TaskNode struct {
	task      Task
	processId string
}

type WorkerPoolConfig struct {
	Concurrency   int
	LogPath      string
	DatabasePath string
}

type WorkerPool struct {
	config    WorkerPoolConfig
	tasksChan chan TaskNode
	wg        sync.WaitGroup
	doneChan  chan struct{}
	state     *state.State
	cancel    context.CancelFunc
	ctx       context.Context
}

func NewWorkerPool(config WorkerPoolConfig) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		config:    config,
		tasksChan: make(chan TaskNode),
		doneChan:  make(chan struct{}),
		state:     &state.State{},
		cancel:    cancel,
		ctx:       ctx,
	}
}

func (wp *WorkerPool) Start() error {
	err := wp.state.ConnectDB(state.ConnectDBParams{
		DBPath: wp.config.DatabasePath,
	})
	if err != nil {
		return err
	}

	for i := 0; i < wp.config.Concurrency; i++ {
		go wp.worker(i)
	}

	return nil
}

func (wp *WorkerPool) AddTask(task Task) (*TaskInfo, error) {
	taskNode := TaskNode{
		task:      task,
		processId: uuid.NewString(),
	}
	wp.wg.Add(1)
	wp.tasksChan <- taskNode
	t, err := wp.state.CreateSingleTask(state.CreateSingleTaskParams{
		ProcessId: taskNode.processId,
	})
	if err != nil {
		return nil, err
	}

	mappedTaskInfo := mapTaskToInfo(*t)

	return &mappedTaskInfo, nil
}

func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.tasksChan)
	wp.wg.Wait()
	close(wp.doneChan)
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}

func (wp *WorkerPool) worker(workerId int) {
	for taskNode := range wp.tasksChan {
		err := wp.setLogPath(wp.config.LogPath)
		if err != nil {
			fmt.Printf("Error setting log path: %v\n", err.Error())
			wp.wg.Done()
			return
		}

		ctx := &ProcessContext{
			WorkerId:  workerId,
			ProcessId: taskNode.processId,
			LogPath:   wp.config.LogPath,
		}

		logFilePath := filepath.Join(wp.config.LogPath, fmt.Sprintf("%s.log", ctx.ProcessId))

		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			_ = ctx.Logger(fmt.Sprintf("Error opening log file: %v", err.Error()))
			wp.wg.Done()
			continue
		}

		ctx.Stdout = file
		ctx.Stderr = file
		ctx.Stdin = file

		_, err = wp.state.TaskQueued(state.TaskQueuedParams{
			ProcessID: taskNode.processId,
			LogPath:   logFilePath,
		})
		if err != nil {
			_ = ctx.Logger(fmt.Sprintf("Error queuing task: %v", err.Error()))
		}

		_ = ctx.Logger("Starting task\n")
		err = ctx.SafeProcess(wp.ctx, taskNode.task)
		if err != nil {
			_ = ctx.Logger(fmt.Sprintf("[%v]: Error processing task: %v\n", ctx.ProcessId, err.Error()))
			_, err = wp.state.TaskFailed(state.TaskFailedParams{
				ProcessID:    ctx.ProcessId,
				ErrorMessage: err.Error(),
			})
			if err != nil {
				_ = ctx.Logger(fmt.Sprintf("Error failing task: %v", err.Error()))
			}
		} else {
			_ = ctx.Logger("Task completed successfully\n")
			_, err = wp.state.TaskCompleted(state.TaskCompletedParams{
				ProcessID: ctx.ProcessId,
			})
			if err != nil {
				_ = ctx.Logger(fmt.Sprintf("Error completing task: %v", err.Error()))
			}
		}
		file.Close()
		wp.wg.Done()
		_ = ctx.Logger(fmt.Sprintf("Worker %d finished\n", workerId))
	}
}

func (wp *WorkerPool) setLogPath(path string) error {
	if path == "" {
		path = "logs"
	}
	wp.config.LogPath = path

	err := os.MkdirAll(wp.config.LogPath, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}


type TaskInfo struct {
	ProcessID  string  `json:"processId"`
	Status     string  `json:"status"`
	LogPath    string  `json:"logPath"`
	WorkerID   *int    `json:"workderId"`
	Error      string  `json:"error"`
	CreatedAt  string  `json:"createdAt"`
	StartedAt  *string `json:"startedAt"`
	FinishedAt *string `json:"finishedAt"`
}

func mapTaskToInfo(task storage.Task) TaskInfo {
	var workerID *int
	if task.WorkerID.Valid {
		val := int(task.WorkerID.Int64)
		workerID = &val
	}

	var startedAt, finishedAt *string
	if task.StartedAt.Valid {
		s := task.StartedAt.Time.Format("2006-01-02 15:04:05")
		startedAt = &s
	}
	if task.FinishedAt.Valid {
		s := task.FinishedAt.Time.Format("2006-01-02 15:04:05")
		finishedAt = &s
	}

	return TaskInfo{
		ProcessID:  task.ProcessID.(string),
		Status:     task.Status,
		LogPath:    task.LogPath.(string),
		WorkerID:   workerID,
		Error:      task.Error.(string),
		CreatedAt:  task.CreatedAt.Format("2006-01-02 15:04:05"),
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
	}
}

func (wp *WorkerPool) GetAllTasks() (*[]TaskInfo, error) {
	raw, err := wp.state.GetAllTasks()
	if err != nil {
		return nil, err
	}

	out := make([]TaskInfo, len(*raw))
	for i, t := range *raw {
		out[i] = mapTaskToInfo(t)
	}

	return &out, nil
}

func (wp *WorkerPool) GetCompletedTasks() (*[]TaskInfo, error) {
	raw, err := wp.state.GetCompletedTasks()
	if err != nil {
		return nil, err
	}

	out := make([]TaskInfo, len(*raw))
	for i, t := range *raw {
		out[i] = mapTaskToInfo(t)
	}

	return &out, nil
}

func (wp *WorkerPool) GetRunningTasks() (*[]TaskInfo, error) {
	raw, err := wp.state.GetRunningTasks()
	if err != nil {
		return nil, err
	}

	out := make([]TaskInfo, len(*raw))
	for i, t := range *raw {
		out[i] = mapTaskToInfo(t)
	}

	return &out, nil
}

func (wp *WorkerPool) GetPendingTasks() (*[]TaskInfo, error) {
	raw, err := wp.state.GetQueuedTasks()
	if err != nil {
		return nil, err
	}

	out := make([]TaskInfo, len(*raw))
	for i, t := range *raw {
		out[i] = mapTaskToInfo(t)
	}

	return &out, nil
}

func (wp *WorkerPool) GetFailedTasks() (*[]TaskInfo, error) {
	raw, err := wp.state.GetFailedTasks()
	if err != nil {
		return nil, err
	}

	out := make([]TaskInfo, len(*raw))
	for i, t := range *raw {
		out[i] = mapTaskToInfo(t)
	}

	return &out, nil
}

type GetTaskByProcessIdParams struct {
	ProcessId string
}

func (wp *WorkerPool) GetTaskByProcessId(params GetTaskByProcessIdParams) (*TaskInfo, error) {
	task, err := wp.state.GetTaskByProcessID(state.GetTaskByProcessIDParams{
		ProcessId: params.ProcessId,
	})
	if err != nil {
		if err.Error() == ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	mappedTaskToInfo := mapTaskToInfo(*task)

	return &mappedTaskToInfo, nil
}
