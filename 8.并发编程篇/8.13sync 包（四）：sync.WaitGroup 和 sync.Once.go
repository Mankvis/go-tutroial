package main

import (
	"fmt"
	"sync"
	"time"
)

/*
在介绍通道的时候，如果启用了多个子协程，我们是这样实现主协程等待子协程执行完毕并退出的：声明一个和子协程数量一致的通道数组，然后为每个子协程分配
一个通道元素，在子协程执行完毕时向对应的通道发送数据；然后在主协程中，我们依次读取这些通道接收子协程发送的数据，只有所有通道都接收到数据才会退出
主协程。
*/

/*
代码看起来是这样的：
*/

//func add(a, b int, ch chan int) {
//	c := a + b
//	fmt.Printf("%d + %d = %d\n", a, b, c)
//	ch <- 1
//}
//
//func main() {
//	chs := make([]chan int, 10)
//	for i := 0; i < 10; i++ {
//		chs[i] = make(chan int)
//		go add(1, i, chs[i])
//	}
//
//	for _, ch := range chs {
//		<-ch
//	}
//
//}

/*
我总感觉这样的实现有点蹩脚，不够优雅，不知道你有没有同感，那有没有更好的实现呢？这就要引入我们今天要讨论的主题：sync 包提供的 sync.WaitGroup
类型。
*/

/*
sync.WaitGroup 类型
*/

/*
sync.WaitGroup 类型是开箱即用的，也是并发安全的。该类型提供了以下三个方法：

Add：WaitGroup 类型有一个计数器，默认值是0，我们可以通过 Add 方法来增加这个计数器的值，通常我们可以通过个方法来标记需要等待的子协程数量；
Done：当某个子协程执行完毕后，可以通过 Done 方法标记已完成，该方法会将所属 WaitGroup 类型实例计数器值减一，通常可以通过 defer 语句来调用它；
Wait：Wait 方法的作用是阻塞当前协程，直到对应 WaitGroup 类型实例的计数器值归零，如果在该方法被调用的时候，对应计数器的值已经是 0，那么它将不
会做任何事情。
*/

/*
至此，你可能已经看出来了，我们完全可以组合使用 sync.WaitGroup 类型提供的方法来替代之前通道中等待子协程执行完毕的实现方法，对应代码如下：
*/

//func addNum(a, b int, deferFunc func()) {
//	defer func() {
//		deferFunc()
//	}()
//	c := a + b
//	fmt.Printf("%d + %d = %d\n", a, b, c)
//}
//
//func main() {
//	var wg sync.WaitGroup
//	wg.Add(10)
//	for i := 0; i < 10; i++ {
//		go addNum(i, 1, wg.Done)
//	}
//	wg.Wait()
//}

/*
看起来代码简洁多了，我们首先在主协程中声明了一个 sync.WaitGroup 类型的 wg 变量，然后调用 Add 方法设置等待子协程数为 10，然后循环启动子协程
，并将 wg.Done 作为 defer 函数传递过去，最后，我们通过 wg.Wait() 等到 sync.WaitGroup 计数器值为 0 时退出程序。
*/

/*
上述代码打印结果和之前通过通道实现的结果是一致的：
0 + 1 = 1
1 + 1 = 2
7 + 1 = 8
4 + 1 = 5
2 + 1 = 3
8 + 1 = 9
3 + 1 = 4
9 + 1 = 10
5 + 1 = 6
6 + 1 = 7
*/

/*
以上就是 sync.WaitGroup 类型的典型使用场景，通过它我们可以轻松实现一主多子的协程协作。需要注意的是，该类型计数器不能小于0，否则会抛出如下
panic：
panic: sync: negative WaitGroup counter
*/

/*
sync.Once 类型
*/

/*
与 sync.WaitGroup 类型类似，sync.Once 类型也是开箱即用和并发安全的，其主要用途是保证指定函数代码只执行一次，类似于单例模式，常用于应用启动
时的一些全局初始化操作。它只提供了一个 Do 方法，该方法只接受一个参数，且这个参数的类型必须是 func()，即无参数无返回值的函数类型。
*/

/*
在具体实现时，sync.Once 还提供了一个 uint32 类型的 done 字段，它的作用是记录 Do 传入函数被调用次数，显然，其对应的值只能是 0 和 1，之所以
设置为 uint32 类型，是为了保证操作的原子性，回想下我们上篇教程中介绍的原子函数，再结合 Do 方法底层实现源码，即可知晓原因，这里不深入探讨了：
*/

/*
func (o *Once) Do(f func()) {
	if atomic.LoadUint32(&o.done) == 1 {
		return
	}
	// Slow-path.
	o.m.Lock()
	defer o.m.Unlock()
	if o.done == 0 {
		defer atomic.StoreUint32(&o.done, 1)
		f()
	}
}
*/

/*
如果 done 字段的值已经是 1 了（通过 atomic.LoadUint32() 原子函数加载），表示该函数已经调用过，否则的话会调用 sync.Once 提供的互斥锁阻塞
其它代码对该类型的访问，然后通过原子操作将 done 的值设置为 1，并调用传入函数。
*/

/*
下面我们通过一个简单的示例来演示 sync.Once 类型的使用：
*/

func doSomething(o *sync.Once) {
	fmt.Printf("Start.")
	o.Do(func() {
		fmt.Println("Do Something...")
	})
	fmt.Println("Finished.")
}

func main() {
	o := &sync.Once{}
	go doSomething(o)
	go doSomething(o)
	time.Sleep(time.Second * 1)
}

/*
上述代码的运行结果是：
Start.
Start.
Do Something...
Finished.
Finished
*/
