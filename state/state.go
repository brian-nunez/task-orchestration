package state

import (
	"context"
	"database/sql"
	_ "embed"
	"sync"

	_ "modernc.org/sqlite"

	"github.com/brian-nunez/task-orchestration/storage"
)

type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusRunning
	TaskStatusCompleted
	TaskStatusFailed
)

func (s TaskStatus) String() string {
	switch s {
	case TaskStatusPending:
		return "pending"
	case TaskStatusRunning:
		return "running"
	case TaskStatusCompleted:
		return "completed"
	case TaskStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

//go:embed schema.sql
var ddl string

type State struct {
	queries *storage.Queries
	dbPath  string
	mu      sync.Mutex
}

type ConnectDBParams struct {
	DBPath string
}

func (s *State) ConnectDB(params ConnectDBParams) error {
	if params.DBPath == "" {
		params.DBPath = ":memory:"
	}

	s.dbPath = params.DBPath

	ctx := context.Background()

	db, err := sql.Open("sqlite", params.DBPath)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `PRAGMA foreign_keys = ON;`)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return err
	}

	s.queries = storage.New(db)

	return nil
}

func (s *State) GetAllTasks() ([]storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	tasks, err := s.queries.GetAllTasks(ctx)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *State) GetQueuedTasks() (*[]storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	tasks, err := s.queries.GetTasksByStatus(ctx, TaskStatusPending.String())
	if err != nil {
		return nil, err
	}

	return &tasks, nil
}

func (s *State) GetRunningTasks() (*[]storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	tasks, err := s.queries.GetTasksByStatus(ctx, TaskStatusRunning.String())
	if err != nil {
		return nil, err
	}

	return &tasks, nil
}

func (s *State) GetCompletedTasks() (*[]storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	tasks, err := s.queries.GetTasksByStatus(ctx, TaskStatusCompleted.String())
	if err != nil {
		return nil, err
	}

	return &tasks, nil
}

func (s *State) GetFailedTasks() (*[]storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	tasks, err := s.queries.GetTasksByStatus(ctx, TaskStatusFailed.String())
	if err != nil {
		return nil, err
	}

	return &tasks, nil
}

type GetTaskByProcessIDParams struct {
	ProcessId string
}

func (s *State) GetTaskByProcessID(params GetTaskByProcessIDParams) (*storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	task, err := s.queries.GetTaskByProcessId(ctx, params.ProcessId)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

type CreateSingleTaskParams struct {
	ProcessId string
}

func (s *State) CreateSingleTask(params CreateSingleTaskParams) (*storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	createdTask, err := s.queries.CreateTask(ctx, storage.CreateTaskParams{
		Type:      "once",
		ProcessID: params.ProcessId,
	})
	if err != nil {
		return nil, err
	}

	return &createdTask, nil
}

func (s *State) UpdateTask(task *storage.Task) (*storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	updatedTask, err := s.queries.UpdateTask(ctx, storage.UpdateTaskParams{
		ID:         task.ID,
		Status:     task.Status,
		LogPath:    task.LogPath,
		Retries:    task.Retries,
		MaxRetries: task.MaxRetries,
		WorkerID:   task.WorkerID,
		ProcessID:  task.ProcessID,
		Error:      task.Error,
		StartedAt:  task.StartedAt,
		FinishedAt: task.FinishedAt,
	})
	if err != nil {
		return nil, err
	}

	return &updatedTask, nil
}

type TaskQueuedParams struct {
	ProcessID string
	LogPath   string
	WorkerID  int
}

func (s *State) TaskQueued(params TaskQueuedParams) (*storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	queuedTask, err := s.queries.QueueTask(ctx, storage.QueueTaskParams{
		LogPath:   params.LogPath,
		ProcessID: params.ProcessID,
		WorkerID: sql.NullInt64{
			Int64: int64(params.WorkerID),
			Valid: params.WorkerID != 0,
		},
	})
	if err != nil {
		return nil, err
	}

	return &queuedTask, nil
}

type TaskCompletedParams struct {
	ProcessID string
}

func (s *State) TaskCompleted(params TaskCompletedParams) (*storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	completedTask, err := s.queries.CompleteTask(ctx, params.ProcessID)
	if err != nil {
		return nil, err
	}

	return &completedTask, nil
}

type TaskFailedParams struct {
	ProcessID    string
	ErrorMessage string
}

func (s *State) TaskFailed(params TaskFailedParams) (*storage.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()

	failedTask, err := s.queries.FailTask(ctx, storage.FailTaskParams{
		ProcessID: params.ProcessID,
		Error:     params.ErrorMessage,
	})
	if err != nil {
		return nil, err
	}

	return &failedTask, nil
}
