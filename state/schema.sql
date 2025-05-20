CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY,
    type TEXT DEFAULT "once", -- once, periodic
    process_id TEXT DEFAULT "",
    log_path TEXT DEFAULT "",
    status TEXT NOT NULL DEFAULT 'pending', -- pending, running, completed, failed
    retries INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    worker_id INT DEFAULT NULL,
    error TEXT DEFAULT "",
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    finished_at TIMESTAMP
);
