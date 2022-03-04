package main

/*
前面我们已经陆续介绍了 sync 包提供的各种同步工具，比如互斥锁、条件变量、原子操作、多协程协作等，今天我们来看另外一种工具。

在高并发场景下，我们会遇到很多问题，垃圾回收（GC）就是其中之一。Go 语言中的垃圾回收是自动执行的，这有利有弊，好处是避免了程序员手动对垃圾进行回
收，简化了代码编写和维护，坏处是垃圾回收的时机无处不在，这在无形之中增加了系统运行时开销。在对系统性能要求较高的高并发场景下，这是我们应该主动去
避免的，因此这需要对对象进行重复利用，以避免产生太多垃圾，而这也就引入了我们今天要讨论的主题 —— sync 包提供的 Pool 类型：
*/

/*
type Pool struct {
	noCopy noCopy

	local     unsafe.Pointer // local fixed-size per-P pool, actual type is [P]poolLocal
	localSize uintptr        // size of the local array

	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() interface{}
}
*/

/*
sync.Pool 是一个临时对象池，可用来临时存储对象，下次使用时从对象池中获取，避免重复创建对象。相应的，该类型提供了 Put 和 Get 方法，分别对临时
对象进行存储和获取。我们可以把 sync.Pool 看作存放可重复使用值的容器，由于 Put 方法支持的参数类型是空接口 interface{}，因此这个值可以是任何
类型，对应的，Get 方法返回值类型也是 interface{}。当我们通过 Get 方法从临时对象池获取临时对象后，会将原来存放在里面的对象删除，最后再返回这
个对象，而如果临时对象池中原来没有存储任何对象，调用 Get 方法时会通过对象池的 New 字段对应函数创建一个新值并返回（这个 New 字段需要在初始化临
时对象池时指定，否则对象池为空时调用 Get 方法返回的可能就是 nil），从而保证无论临时对象池中是否存在值，始终都能返回结果。
*/

/*
下面我们来看个简单的示例代码：
*/

//func main() {
//	var pool = &sync.Pool{
//		New: func() interface{} {
//			return "Hello,World!"
//		},
//	}
//	value := "Hello,学员君!"
//	pool.Put(value)
//	fmt.Println(pool.Get())
//	fmt.Println(pool.Get())
//}

/*
在这段代码中，我们首先声明并初始化了一个临时对象池 pool，并定义了 New 字段，这是一个 func() interface{} 类型的函数，然后我们通过 Put 方法
存储一个字符串对象到 pool，最后通过 Get 方法获取该对象并打印，当我们再次在 pool 实例上调用 Get 方法时，会发现存储的字符串已经不存在，而是通
过 New 字段对应函数返回的字符串对象。
*/

/*
此外，我们还可以利用 sync.Pool 的特性在多协程并发执行场景下实现对象的复用，因为 sync.Pool 本身是并发安全地，我们可以在程序开始执行时全局唯一
初始化 Pool 对象，然后在并发执行的协程之间通过这个临时对象池来存储和获取对象：
*/

//func testPut(pool *sync.Pool, deferFunc func()) {
//	defer func() {
//		deferFunc()
//	}()
//	value := "Hello,学员君!"
//	pool.Put(value)
//}
//
//func main() {
//	var wg sync.WaitGroup
//	wg.Add(1)
//	var pool = &sync.Pool{
//		New: func() interface{} {
//			return "Hello,World!"
//		},
//	}
//	go testPut(pool, wg.Done)
//	wg.Wait()
//	fmt.Println(pool.Get())
//}

/*
我们日常使用较多的 fmt.Println 其实也是基于 sync.Pool 实现 printer 对象在并发调用时的重复使用的，printer 对象可用于打印对象，其初始化和
释放代码如下所示：
*/

/*
func newPrinter() *pp {
    p := ppFree.Get().(*pp)
    p.panicking = false
    p.erroring = false
    p.fmt.init(&p.buf)
    return p
}

func (p *pp) free() {
    if cap(p.buf) > 64<<10 {
        return
    }

    p.buf = p.buf[:0]
    p.arg = nil
    p.value = reflect.Value{}
    ppFree.Put(p)
}
*/

/*
而 ppFree 正是 sync.Pool 对象实例：
*/

/*
var ppFree = sync.Pool{
    New: func() interface{} { return new(pp) },
}
*/

/*
通过以上示例可以看到，临时对象池 sync.Pool 非常适用于在并发编程中用作临时对象缓存，实现对象的重复使用，优化 GC，提升系统性能，但是由于不能设置
对象池大小，而且放进对象池的临时对象每次 GC 运行时会被清除，所以只能用作简单的临时对象池，不能用作持久化的长连接池，比如数据库连接池、Redis
连接池。
*/