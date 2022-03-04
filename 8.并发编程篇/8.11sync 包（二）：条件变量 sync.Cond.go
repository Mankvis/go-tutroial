package main

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

/*
sync 包还提供了一个条件变量类型 sync.Cond，它可以和互斥锁或读写锁（以下统称互斥锁）组合使用，用来协调想要访问共享资源的线程。
*/

/*
不过，与互斥锁不同，条件变量 sync.Cond 的主要作用并不是保证在同一时刻仅有一个线程访问某一个共享资源，而是在对应的共享资源状态发生变化时，通知
其它因此而阻塞的线程。条件变量总是和互斥锁组合使用，互斥锁为共享资源的访问提供互斥支持，而条件变量可以就共享资源的状态变化向相关线程发出通知，重
在「协调」。

下面，我们来看看如何使用条件变量 sync.Cond。
*/

/*
sync.Cond 是一个结构体：
type Cond struct {

  noCopy noCopy

  // L is held while observing or changing the condition
  L Locker

  notify  notifyList
  checker copyChecker
}
*/

/*
提供了三个方法：

// 等待通知
func (c *Cond) Wait() {
  c.checker.check()
  t := runtime_notifyListAdd(&c.notify)
  c.L.Unlock()
  runtime_notifyListWait(&c.notify, t)
  c.L.Lock()
}

// 单发通知
func (c *Cond) Signal() {
  c.checker.check()
  runtime_notifyListNotifyOne(&c.notify)
}

// 广播通知
func (c *Cond) Broadcast() {
  c.checker.check()
  runtime_notifyListNotifyAll(&c.notify)
}
*/

/*
我们可以通过 sync.NewCond 返回对应的条件变量实例，初始化的时候需要传入互斥锁，该互斥锁实例会赋值给 sync.Cond 的 L 属性：
*/

//locker := &sync.Mutex{}
//cond := sync.NewCond(locker)

/*
sync.Cond 主要实现一个条件变量，假设 goroutine A 执行前需要等待另外一个 goroutine B 的通知，那么处于等待状态的 goroutine A 会保存在一
个通知列表，也就是说需要某种变量状态的 goroutine A 将会等待（Wait）在那里，当某个时刻变量状态改变时，负责通知的 goroutine B 会通过对条件变
量通知的方式（Broadcast/Signal）来通知处于等待条件变量的 goroutine A，这样就可以在共享内存中实现类似「消息通知」的同步机制。
*/

/*
下面来看一个具体的示例。假设我们有一个读取器和一个写入器，读取器必须依赖写入器对缓冲区进行数据写入后，才可以从缓冲区中读取数据，写入器每次完成写
入数据后，都需要通过某种通知机制通知处于阻塞状态的读取器，告诉它可以对数据进行访问，这种场景正好可以通过条件变量来实现：
*/

// 数据bucket
//type DataBucket struct {
//	buffer *bytes.Buffer // 缓冲区
//	mutex  *sync.RWMutex // 互斥区
//	cond   *sync.Cond    // 条件变量
//}
//
//func NewDataBucket() *DataBucket {
//	buf := make([]byte, 0)
//	db := &DataBucket{
//		buffer: bytes.NewBuffer(buf),
//		mutex:  new(sync.RWMutex),
//	}
//	db.cond = sync.NewCond(db.mutex.RLocker())
//	return db
//}
//
//// Read 读取器
//func (db *DataBucket) Read(i int) {
//	db.mutex.RLock()         // 打开读锁
//	defer db.mutex.RUnlock() // 结束后释放读锁
//	var data []byte
//	var d byte
//	var err error
//	for {
//		// 每次读取一个字节
//		if d, err = db.buffer.ReadByte(); err != nil {
//			if err == io.EOF { // 缓冲区数据为空时执行
//				if string(data) != "" { // data不为空,则打印它
//					fmt.Printf("reader-%d: %s\n", i, data)
//				}
//				db.cond.Wait()  // 缓冲区为空,通过wait方法等待通知,进入阻塞状态
//				data = data[:0] // 将data清空
//				continue
//			}
//		}
//		data = append(data, d) // 将读取到到数据添加到data中
//	}
//}
//
//// Put 写入器
//func (db *DataBucket) Put(d []byte) (int, error) {
//	db.mutex.Lock()         // 打开写锁
//	defer db.mutex.Unlock() // 结束后释放写锁
//	// 写入一个数据块
//	n, err := db.buffer.Write(d)
//	db.cond.Signal() // 写入数据后通过Signal通知处于阻塞状态到读取器
//	return n, err
//}
//
//func main() {
//	db := NewDataBucket()
//	go db.Read(1) // 开启读取器协程
//	go func(i int) {
//		d := fmt.Sprintf("data-%d", i)
//		db.Put([]byte(d)) // 写入数据到缓冲区
//	}(1) // 开启写入器协程
//	time.Sleep(100 * time.Millisecond)
//}

/*
这里我们使用了读写互斥锁，在读取器里面使用读锁，在写入器里面使用写锁，并且通过 defer 语句释放锁，然后在锁保护的情况下，通过条件变量协调读写线
程：在读线程中，当缓冲区为空的时候，通过 db.cond.Wait() 阻塞读线程；在写线程中，当缓冲区写入数据的时候通过 db.cond.Signal() 通知读线程继
续读取数据。
*/

/*
执行上述示例代码，结果如下：
reader-1: data-1
*/

/*
上述示例代码只有一个读取器，一个写入器，如果都有多个呢？我们可以通过启动多个读写协程来模拟，此外，通知单个阻塞线程用 Signal 方法，通知多个阻塞
线程需要使用 Broadcast 方法，按照这个思路，我们来改写上述示例代码如下：
*/

type DataBucket struct {
	buffer *bytes.Buffer // 缓冲区
	mutex  *sync.RWMutex // 互斥区
	cond   *sync.Cond    // 条件变量
}

func NewDataBucket() *DataBucket {
	buf := make([]byte, 0)
	db := &DataBucket{
		buffer: bytes.NewBuffer(buf),
		mutex:  new(sync.RWMutex),
	}
	db.cond = sync.NewCond(db.mutex.RLocker())
	return db
}

// Read 读取器
func (db *DataBucket) Read(i int) {
	db.mutex.RLock()         // 打开读锁
	defer db.mutex.RUnlock() // 结束后释放读锁
	var data []byte
	var d byte
	var err error
	for {
		// 每次读取一个字节
		if d, err = db.buffer.ReadByte(); err != nil {
			if err == io.EOF { // 缓冲区数据为空时执行
				if string(data) != "" { // data不为空,则打印它
					fmt.Printf("reader-%d: %s\n", i, data)
				}
				db.cond.Wait()  // 缓冲区为空,通过wait方法等待通知,进入阻塞状态
				data = data[:0] // 将data清空
				continue
			}
		}
		data = append(data, d) // 将读取到到数据添加到data中
	}
}

// Put 写入器
func (db *DataBucket) Put(d []byte) (int, error) {
	db.mutex.Lock()         // 打开写锁
	defer db.mutex.Unlock() // 结束后释放写锁
	// 写入一个数据块
	n, err := db.buffer.Write(d)
	db.cond.Broadcast() // 写入数据后通过Broadcast通知处于阻塞状态到读取器
	return n, err
}

func main() {
	db := NewDataBucket()
	for i := 0; i < 3; i++ { // 启动多个读取器
		go db.Read(i)
	}

	for j := 0; j < 10; j++ { // 启动多个写入器
		go func(i int) {
			d := fmt.Sprintf("data-%d", i)
			db.Put([]byte(d)) // 写入数据到缓冲区
		}(j) // 开启写入器协程
		time.Sleep(100 * time.Millisecond) // 每次启动一个写入器暂停100ms,让读取器阻塞
	}

}

/*
执行上述代码，打印结果如下：
reader-2: data-0
reader-0: data-1
reader-2: data-2
reader-1: data-3
reader-2: data-4
reader-2: data-5
reader-2: data-6
reader-2: data-7
reader-0: data-8
reader-1: data-9
*/

/*
可以看到，通过互斥锁+条件变量，我们可以非常方便的实现多个 Go 协程之间的通信，但是这个还是比不上 channel，因为 channel 还可以实现数据传递，
条件变量只是发送信号，唤醒被阻塞的协程继续执行，另外 channel 还有超时机制，不会出现协程等不到信号一直阻塞造成内存堆积问题，换句话说，channel
可以让程序更可控。
*/
