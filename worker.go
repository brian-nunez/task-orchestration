package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/brian-nunez/task-orchestration/state"
	"github.com/brian-nunez/task-orchestration/storage"
	"github.com/google/uuid"
)

type Task interface {
	Process(taskContext *ProcessContext) error
}

type TaskNode struct {
	task      Task
	processId string
}

type WorkerPool struct {
	Concurreny   int
	LogPath      string
	tasksChan    chan TaskNode
	wg           sync.WaitGroup
	doneChan     chan struct{}
	DatabasePath string
	state        *state.State
}

func (wp *WorkerPool) Start() error {
	wp.tasksChan = make(chan TaskNode)
	wp.doneChan = make(chan struct{})

	wp.state = &state.State{}
	err := wp.state.ConnectDB(state.ConnectDBParams{
		DBPath: wp.DatabasePath,
	})
	if err != nil {
		return err
	}

	for i := 0; i < wp.Concurreny; i++ {
		go wp.worker(i)
	}

	return nil
}

func (wp *WorkerPool) AddTask(task Task) {
	taskNode := TaskNode{
		task:      task,
		processId: uuid.NewString(),
	}
	wp.wg.Add(1)
	wp.tasksChan <- taskNode
	wp.state.CreateSingleTask(state.CreateSingleTaskParams{
		ProcessId: taskNode.processId,
	})
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
	for taskNode := range wp.tasksChan {
		err := wp.setLogPath(wp.LogPath)
		if err != nil {
			fmt.Printf("Error setting log path: %v\n", err.Error())
			wp.wg.Done()
			return
		}

		ctx := &ProcessContext{
			WorkerId:  workerId,
			ProcessId: taskNode.processId,
			LogPath:   wp.LogPath,
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

		wp.state.TaskQueued(state.TaskQueuedParams{
			ProcessID: taskNode.processId,
			LogPath:   filepath,
		})

		ctx.Logger("Starting task\n")
		err = ctx.SafeProcess(taskNode.task)
		// err = taskNode.task.Process(ctx)
		if err != nil {
			ctx.Logger(fmt.Sprintf("[%v]: Error processing task: %v\n", ctx.ProcessId, err.Error()))
			wp.state.TaskFailed(state.TaskFailedParams{
				ProcessID:    ctx.ProcessId,
				ErrorMessage: err.Error(),
			})
		} else {
			ctx.Logger("Task completed successfully\n")
			wp.state.TaskCompleted(state.TaskCompletedParams{
				ProcessID: ctx.ProcessId,
			})
		}
		file.Close()
		wp.wg.Done()
		ctx.Logger(fmt.Sprintf("Worker %d finished\n", workerId))
	}
}

func (wp *WorkerPool) setLogPath(path string) error {
	if path != "" {
		wp.LogPath = path
	} else {
		wp.LogPath = "logs"
	}

	err := os.MkdirAll(wp.LogPath, os.ModePerm)
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
