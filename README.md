# 🧵 Task Orchestration System for Go CLI Tools

🛠️ A simple, clean worker pool and task queue system designed for CLI apps.

Built in Go. Logs every task. Handles concurrency like a champ.

---

## 📦 Installation

Install directly using go get:

```sh
go get github.com/brian-nunez/task-orchestration@latest
```

Import it into your project:

```go
import (
    worker "github.com/brian-nunez/task-orchestration"
)
```

---

## 🚀 Usage

Set up a worker pool and submit tasks:

```go
pool := &worker.WorkerPool{
	Concurreny: 4,
	LogPath:    "logs",
}

pool.Start()
defer pool.Stop()

pool.AddTask(&tasks.LoggerTask{
	Text:     "Doing some work",
	LogLevel: "INFO",
	Delay:    time.Second,
})

pool.Wait()
```

This will:

* Run tasks in parallel across 4 workers
* Write logs to individual .log files under the logs/ directory

---

## 🛠️ Define Your Own Task

Just implement the Task interface:

```go
type Task interface {
	Process(taskContext *worker.ProcessContext) error
}
```

Here’s a simple example:

```go
type PrintTask struct {
	Message string
}

func (t *PrintTask) Process(ctx *worker.ProcessContext) error {
	ctx.Logger("Processing task")
	fmt.Println(t.Message)
	return nil
}
```

---

## 📄 Log Output

Each task creates a log file:

```txt
logs/{process-id}.log
```

These logs include:

* Worker ID
* Custom messages

---

## ⚙️ Features

* ✅ Plug-and-play with your Go CLI
* ✅ Per-task log files
* ✅ Easy to extend with your own tasks

---

## ✨ License

See LICENSE file for details.

---

## Authors

- [brian-nunez](https://www.github.com/brian-nunez) - Maintainer
