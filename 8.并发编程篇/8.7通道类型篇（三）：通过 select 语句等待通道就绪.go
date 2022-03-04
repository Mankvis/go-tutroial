package main

import (
	"fmt"
	"math/rand"
	"time"
)

/*
Go 语言还支持通过 select 分支语句选择指定分支代码执行，select 语句和之前介绍的 switch 语句语法结构类似，不同之处在于 select 的每个 case
语句必须是一个通道操作，要么是发送数据到通道，要么是从通道接收数据，此外 select 语句也支持 default 分支：
*/

//select {
//case <-cha1:
//	// 如果从chan1通道成功接收数据,则执行该分支代码
//case chan2<-1:
//	// 如果成功向chan2通道成功发送时数据,则执行该分支代码
//default:
//	// 如果上面都没有成功,则进入default分之处理流程
//}

/*
注：Go 语言的 select 语句借鉴自 Unix 的 select() 函数，在 Unix 中，可以通过调用 select() 函数来监控一系列的文件句柄，一旦其中一个文件句
柄发生了 IO 动作，该 select() 调用就会被返回（C 语言中就是这么做的），后来该机制也被用于实现高并发的 Socket 服务器程序。Go 语言直接在语言级
别支持 select 关键字，用于处理并发编程中通道之间异步 IO 通信问题。
*/

/*
可以看出，select 不像 switch，case 后面并不带判断条件，而是直接去查看 case 语句，每个 case 语句都必须是一个面向通道的操作，比如上面的示例
代码中，第一个 case 试图从 chan1 接收数据并直接忽略读到的数据，第二个 case 试图向 chan2 通道发送一个整型数据 1，需要注意的是这两个 case
的执行不是 if...else... 那种先后关系，而是会并发执行，然后 select 会选择先操作成功返回的那个 case 分支去执行，如果两者同时返回，则随机选择
一个执行，如果这两者都没有返回，则进入 default 分支，这里也不会出现阻塞，如果 chan1 通道为空，或者 chan2 通道已满，就会立即进入 default
分支，但是如果没有 default 语句，则会阻塞直到某个通道操作成功。
*/

/*
因此，借助 select 语句我们可以在一个协程中同时等待多个通道达到就绪状态：
图片: https://laravel.gstatics.cn/wp-content/uploads/2019/08/20039d6035200f3e610b556c33ef63f2.jpg
*/

/*
这些通道操作是并发的，任何一个操作成功，就会进入该分支执行代码，否则程序就会处于挂起状态，如果要实现非阻塞操作，可以引入 default 语句。
*/

/*
下面我们基于 select 语句来实现一个简单的示例代码：
*/

//func main() {
//	chs := [3]chan int{
//		make(chan int, 1),
//		make(chan int, 1),
//		make(chan int, 1),
//	}
//
//	// 随机生成0-2之间的数字
//	rand.Seed(time.Now().UnixNano())
//	index := rand.Intn(3)
//	fmt.Printf("随机索引/数值: %d\n", index)
//
//	chs[index] <- index
//
//	// 哪一个通道中有值,哪个对应的分之就会执行
//	select {
//	case <-chs[0]:
//		fmt.Println("第一个条件分支被选中")
//	case <-chs[1]:
//		fmt.Println("第二个条件分支被选中")
//	case <-chs[2]:
//		fmt.Println("第三个条件分支被选中")
//	default:
//		fmt.Println("没有分支被选中")
//	}
//}

/*
在这段代码中，我们创建了一个包含 3 个 chan int 类型元素的通道数组，然后随机往某个通道中发送一个随机数据，再通过 select 语句从上面定义的三个
通道中接收数据，只要是发送数据成功，就一定能将其取出来，如果通道都为空，则直接执行 default 语句。
*/

/*
执行上述这段代码打印结果如下：
随机索引/数值: 2
第三个条件分支被选中: 2
*/

/*
如果我们将 chs[index] <- index 这一行注释掉，则打印结果如下：
随机索引/数值: 2
没有分支被选中
*/

/*
另外，需要注意的是，select 语句只能对其中的每一个 case 表达式各求值一次，如果我们想连续操作其中的通道的话，需要通过在 for 语句中嵌入 select
语句的方式来实现：
*/

func main() {
	chs := [3]chan int{
		make(chan int, 3),
		make(chan int, 3),
		make(chan int, 3),
	}
	rand.Seed(time.Now().UnixNano())
	// 随机生成0-2的数字
	index1 := rand.Intn(3)
	fmt.Printf("随机索引/数值: %d\n", index1)
	// 向通道发送随机数字
	chs[index1] <- rand.Int()

	index2 := rand.Intn(3)
	fmt.Printf("随机索引/数值: %d\n", index2)
	chs[index2] <- rand.Int()
	index3 := rand.Intn(3)
	fmt.Printf("随机索引/数值: %d\n", index3)
	chs[index3] <- rand.Int()

	//哪个通道中有值,哪个对应的分支就会执行
	for i := 0; i < 3; i++ {
		select {
		case num, ok := <-chs[0]:
			if !ok {
				break
			}
			fmt.Println("第一个条件分支被执行: chs[0] =>", num)
		case num, ok := <-chs[1]:
			if !ok {
				break
			}
			fmt.Println("第二个条件分支被执行: chs[1] =>", num)
		case num, ok := <-chs[2]:
			if !ok {
				break
			}
			fmt.Println("第三个条件分支被执行: chs[2] =>", num)
		default:
			fmt.Println("没有分支被执行")
		}
	}
}

/*
但这时要注意，简单地在 select 语句的分支中使用 break 语句，只能结束当前的 select 语句的执行，而并不会对外层的 for 语句产生作用，如果 for
循环本身没有退出机制的话会无休止地运行下去。
*/
