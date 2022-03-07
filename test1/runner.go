package main

import (
	"fmt"
	"go-tutorial/test1/runner"
	"log"
	"os"
	"time"
)

const timeout = time.Second * 10

func main() {
	fmt.Println("开始运行...")

	// 初始化作业运行器
	r := runner.New(timeout)

	// 调度三个后台处理任务
	r.Add(createTask(), createTask(), createTask())

	// 启动作业运行器
	if err := r.Start(); err != nil {
		switch err {
		case runner.ErrTimeout:
			log.Printf("作业程序因运行超时而中止.")
			os.Exit(1)
		case runner.ErrInterrupt:
			log.Printf("作业程序因系统发生中断时间而终止")
			os.Exit(2)
		}
	}

}

func createTask() func(int) {
	return func(id int) {
		log.Printf("Processor - Task #%d.\n", id)
		time.Sleep(time.Duration(id) * time.Second)
	}
}
