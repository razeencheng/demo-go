# [Go学习笔记(二) | 我对recover的一点误解](https://razeen.me/post/daily-go-recover.html)

在golang的官方介绍中是这么介绍**Recover**函数的。

<!--more-->

```doc
Recover is a built-in function that regains control of a panicking goroutine. Recover is only useful inside deferred functions. During normal execution, a call to recover will return nil and have no other effect. If the current goroutine is panicking, a call to recover will capture the value given to panic and resume normal execution.
```

也就是说当一个协程发生panic时，recover函数会捕捉到panic同时恢复正常的顺序。



像如下这样的代码：

``` golang
package main

import (
	"fmt"
	"log"
	"time"
)

func main() {

	ch := make(chan int, 10)

	for i := 2; i > 0; i-- {
		go func(i int) {
			defer func() {
				err := recover()
				if err != nil {
					log.Println(err)
				}
			}()
			for val := range ch {
				fmt.Println("---->", val, "Go", i)
				if val%2 == 1 && i == 1 {
					panic("BOOM BOOM")
				}
				time.Sleep(2 * time.Second)
			}
		}(i)
	}

	var i int
	for {
		ch <- i
		time.Sleep(1 * time.Second)
		fmt.Println(i, "<---")
		i++
	}
}
```

我们向通道中写入，两个协成在接收，同时设计一个协程在一定的时候panic。那么在panic后，再recover，那在该协程的管道还能接收么？

过去，我一直以为，recover函数会重新恢复，该协程会类似重启一般==。 

```bash
$ go run recover.go
----> 0 Go 2
0 <---
----> 1 Go 1
2018/01/21 21:52:54 BOOM BOOM
1 <---
----> 2 Go 2
2 <---
3 <---
----> 3 Go 2
4 <---
----> 4 Go 2
5 <---
6 <---
----> 5 Go 2
7 <---
8 <---
----> 6 Go 2
...
```

但从执行的结果看：

​	只有一个协程在工作，另外一个还是挂了。这时，我才意识到，recover函数并没有恢复原有的协程。只是当该协程panic后会执行defer。而在defer中，recover函数将panic拦截下来了，不会向外面抛出，从而导致其他协程的执行并不受到影响。但，已经panic的协程还是挂了。