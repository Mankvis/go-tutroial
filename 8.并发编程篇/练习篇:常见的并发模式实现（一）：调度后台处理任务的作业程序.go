package main

/*
代码为 root目录下 -> test1
*/

/*
关于 Go 语言并发编程基本概念和 Goroutine、Channel 以及锁机制的使用，学院君在 Go 入门教程并发编程章节已经详细介绍过了，这里主要演示通过并发
编程在 Go 程序中实现一些常见的并发模式。
*/

/*
首先，我们来看如何开发需要调用后台处理任务的程序，这个程序可能会作为 Cron 作业执行，或者在基于定时任务的云环境（iron.io）里执行。
*/

/*
我们创建一个 runner 包，在该包中创建一个 job.go 文件，编写对应的作业类实现代码如下：
*/

/*
package runner

import (
    "errors"
    "os"
    "os/signal"
    "time"
)

type JobRunner struct {
    interrupt chan os.Signal
    complete chan error
    timeout <-chan time.Time
    tasks []func(int)
}

var ErrTimeout = errors.New("received timeout")
var ErrInterrupt = errors.New("received interrupt")

func New(d time.Duration) *JobRunner {
    return &JobRunner{
        interrupt: make(chan os.Signal, 1),
        complete:  make(chan error),
        timeout:   time.After(d),
    }
}

func (r *JobRunner) Add(tasks ...func(int)) {
    r.tasks = append(r.tasks, tasks...)
}

func (r *JobRunner) Start() error {
    // 接收系统中断信号通知
    signal.Notify(r.interrupt, os.Interrupt)

    go func() {
        r.complete <- r.run()
    }()

    select {
    case err := <-r.complete:
        return err
    case <-r.timeout:
        return ErrTimeout
    }
}

func (r *JobRunner) run() error {
    for id, task := range r.tasks {
        if r.gotInterrupt() {
            return ErrInterrupt
        }
        task(id)
    }
    return nil
}

func (r *JobRunner) gotInterrupt() bool {
    select {
    case <-r.interrupt:
        signal.Stop(r.interrupt)
        return true
    default:
        return false
    }
}
*/

/*
上述代码展示了根据调度运行的、无人值守的、面向任务的并发模式程序：调用 Start() 方法启动作业运行器后，会通过协程异步运行作业中的所有后台处理任
务，然后通过 select 选择语句判定作业程序是运行结束正常退出、还是收到系统中断信号退出、亦或是超时异常退出，如果正常退出，返回的状态码是 nil，否
则是非空的错误值。
*/

/*
这样一来，不管后台处理任务有多少个、耗时多久，都可以做到并发运行，从而提升程序性能和运行效率。
*/

/*
我们可以编写一个入口程序 runner.go 来调用上述调度后台处理任务的作业程序：
*/

/*
package main

import (
    "fmt"
    "log"
    "os"
    "test/runner"
    "time"
)

const timeout = 3 * time.Second

func main()  {
    fmt.Println("开始运行...")

    // 初始化作业运行器
    r := runner.New(timeout)

    // 调度三个后台处理任务
    r.Add(createTask(), createTask(), createTask())

    // 启动作业运行器
    if err := r.Start(); err != nil {
        switch err {
        case runner.ErrTimeout:
            log.Println("作业程序因运行超时而终止")
            os.Exit(1)
        case runner.ErrInterrupt:
            log.Println("作业程序因系统发生中断事件而终止")
            os.Exit(2)
        }
    }
}

// 编写一个模拟后台处理任务
func createTask() func(int) {
    return func(id int) {
        log.Printf("Processor - Task #%d.", id)
        time.Sleep(time.Duration(id) * time.Second)
    }
}
*/

/*
附：上述示例代码目录结构如下（go.mod 中的 package 名称是 test）：
*/

/*
--go (项目根目录 ~/Development/go)
  |--src
      |--test
          |--runner
              |--job.go
          |--runner.go
          |--go.mod
*/

/*
运行上述代码，输出结果如下：
图片地址: https://laravel.gstatics.cn/storage/uploads/images/gallery/2020-10/16017391023667.jpg
*/

/*
由于系统超时时间是 3s，而后台处理任务总耗时是3s，因此程序整体运行时间是超过 3s 的，所以显示超时退出，如果我们将系统超时时间延长至 5s，则会正常
退出。
*/
