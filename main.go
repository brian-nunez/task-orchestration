package main

import (
	"github.com/brian-nunez/task-orchestration/tasks"
	"github.com/brian-nunez/task-orchestration/worker"
)

func main() {
	logList := []worker.Task{
		&tasks.M3U8Task{
			URL:    "https://mcloud.vvid30c.site/hls/f397c5a266bae5ef-a69f0b7e8469ba14f40bf65d-9546d5addfca0e73bf4fbf6287b9069e692ed523bb2f01e512a8ee8201f716644327b2d4500cf3c5fab8ceae04/master.m3u8",
			Output: "southpark-s26_e01.mp4",
		},
		// &tasks.M3U8Task{
		// 	URL:    "https://mcloud.vvid30c.site/hls/29b5b0bad1fe5851-058c6b037308ce99c3ae9ed2-20f0ec6d4d69bca18fde6bafc10a18ee0c83e7c2c3320c8ce351bcc739e7280738d1ea333265714bd5667fd4f9/master.m3u8",
		// 	Output: "southpark-s26_e02.mp4",
		// },
		// &tasks.LoggerTask{
		// 	Text:     "wow",
		// 	Delay:    time.Second * 2,
		// 	LogLevel: "INFO",
		// },
		// &tasks.LoggerTask{
		// 	Text:     "wow1",
		// 	Delay:    time.Second * 3,
		// 	LogLevel: "DEBUG",
		// },
		// &tasks.LoggerTask{
		// 	Text:     "wow2",
		// 	Delay:    time.Second * 1,
		// 	LogLevel: "INFO",
		// },
		// &tasks.LoggerTask{
		// 	Text:     "wow3",
		// 	Delay:    time.Second * 2,
		// 	LogLevel: "INFO",
		// },
		// &tasks.LoggerTask{
		// 	Text:     "wow4",
		// 	Delay:    time.Second * 1,
		// 	LogLevel: "INFO",
		// },
		// &tasks.LoggerTask{
		// 	Text:     "wow5",
		// 	Delay:    time.Second * 8,
		// 	LogLevel: "INFO",
		// },
		// &tasks.LoggerTask{
		// 	Text:     "wow6",
		// 	Delay:    time.Second * 3,
		// 	LogLevel: "INFO",
		// },
	}

	pool := worker.WorkerPool{
		Tasks:      logList,
		Concurreny: 1,
	}

	pool.Run()
}
