-- name: GetAllTasks :many
SELECT * FROM tasks;

-- name: GetTasksByStatus :many
SELECT * FROM tasks WHERE status = ? ORDER BY updated_at DESC;

-- name: GetTaskByProcessId :one
SELECT * FROM tasks WHERE process_id = ?;

-- name: CreateTask :one
INSERT INTO tasks ("type", "process_id", "status") VALUES (?, ?, "pending") RETURNING *;

-- name: UpdateTask :one
UPDATE tasks SET
    status = ?,
    log_path = ?,
    retries = ?,
    max_retries = ?,
    worker_id = ?,
    process_id = ?,
    error = ?,
    updated_at = CURRENT_TIMESTAMP,
    started_at = ?,
    finished_at = ?
WHERE id = ?
RETURNING *;

-- name: QueueTask :one
UPDATE tasks SET
    status = 'running',
    log_path = ?,
    worker_id = ?,
    updated_at = CURRENT_TIMESTAMP,
    started_at = CURRENT_TIMESTAMP
WHERE process_id = ?
RETURNING *;

-- name: CompleteTask :one
UPDATE tasks SET
    status = 'completed',
    updated_at = CURRENT_TIMESTAMP,
    finished_at = CURRENT_TIMESTAMP
WHERE process_id = ?
RETURNING *;

-- name: FailTask :one
UPDATE tasks SET
    status = 'failed',
    updated_at = CURRENT_TIMESTAMP,
    finished_at = CURRENT_TIMESTAMP,
    error = ?
WHERE process_id = ?
RETURNING *;
