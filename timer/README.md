# [Go学习笔记（九） 计时器的生命周期[译]](https://razeencheng.com/post/go-timers-life-cycle.html)



![Illustration created for “A Journey With Go”, made from the original Go Gopher, created by Renee French.](https://st.razeen.cn/image/go-timer.png)



*全文基于GO 1.14*



计时器在定时执行一些任务时很有用。Go内部依靠调度器来管理创建的计时器。而Go的调度程序是协作式的调度方式，这会让整个调度看起来比较复杂，因为goroutune必须自己停止（依赖channel阻塞或system call), 或者由调度器自己在某个调度点暂停。

<!--more-->



*有关抢占的更多信息，建议您阅读作者的文章[Go: Goroutine and Preemption](https://medium.com/a-journey-with-go/go-goroutine-and-preemption-d6bc2aa2f4b7)*.



### 生命周期

下面是一段简单示例代码：

```go
func main(){
	sigs := make(chan os.Signal,1)
	signal.Notify(sigs,syscall.SIGINT,syscall.SIGTERM)

	time.AfterFunc(time.Second, func() {
		println("done")
	})

	<- sigs
}
```

计时器创建后，他会保存到一个链接到当前P的计时器内部列表上，下图就是这段代码的表示形式：

![](https://st.razeen.cn/image/timer-on-p.png)

*有关G，M，P模型的更多信息，建议您阅读作者的文章[Go: Goroutine, OS Thread and CPU Management](https://medium.com/a-journey-with-go/go-goroutine-os-thread-and-cpu-management-2f5a5eaf518a)*





从图中可以看到，一旦创建了计时器，它就会注册一个内部回调，该内部回调将使用`go`回调用户函数，并将其转换为goroutine。



然后，将通过调度程序管理计时器。在每一轮调度中，它都会检查计时器是否准备好运行，如果准备就绪，则准备运行。实际上，由于Go调度程序本身不会运行任何代码，因此运行计时器的回调会将其goroutine排队到本地队列中。然后，当调度程序在队列中将其接收时，goroutine将运行。如选图所示：

![](https://st.razeen.cn/image/timer-on-p2.png)

根据本地队列的大小，计时器运行可能会稍有延迟。不过，由于Go 1.14中的异步抢占，goroutines在运行时间10ms后会被抢占，降低了延迟的可能性。



###  延迟？

为了了解计时器可能存在的延迟，我们创造一个场景：从同一goroutine创建大量计时器。

由于计时器都链接到当前`P`，因此繁忙的`P`无法及时运行其链接的计时器。代码如下：

``` go
func main(){

	trace.Start(os.Stderr)
	defer trace.Stop()

	sigs := make(chan os.Signal,1)
	signal.Notify(sigs,syscall.SIGINT,syscall.SIGTERM)

	//time.AfterFunc(time.Second, func() {
	//	println("done")
	//})


	var num int64

	for i:=0; i< 1e3 ; i++ {
		time.AfterFunc(time.Second, func() {
			atomic.AddInt64(&num,1)
		})
	}

	t:= 0
	for i:=0;i<1e10; i++ {
		t ++
	}
	_ = t

	<- sigs

	println(num,"timers created,",t,"iterations done")
}
```

通过go tool trace， 我们可以看到goroutine正在占用处理器：

![](https://st.razeen.cn/image/timer-on-p3.png)

由于异步抢占的原因，代表正在运行的goroutine图形被分成了大量较小的块。



在这些块中，一个空间看起来比其他空间大。让我们放大一下：

![](https://st.razeen.cn/image/timer-on-p4.png)



在该计时器需要运行时，就会发生改情况。此时，当前goroutine已被Go调度程序抢占和取代。调度程序将计时器转换为可运行的goroutine，如图所示。

但是，当前线程的Go调度程序并不是唯一运行计时器的调度程序。Go实施了一种计时器窃取策略，以确保在当前P繁忙时可以由另一个P运行计时器。由于异步抢占，它不太可能发生，但是在我们的示例中，由于使用了大量的计时器，它发生了。如下图所示：

![](https://st.razeen.cn/image/timer-on-p5.png)



如果我们不考虑计时器窃取，将发生以下情况：

![](https://st.razeen.cn/image/timer-on-p6.png)



持有计时器的所有goroutine都会添加到本地队列中。然后，由于 `P`之间的窃取，将准确的调度计时器。

所以，由于异步抢占和工作窃取，延迟几乎不可能发生。



> 原文 [Go: Timers’ Life Cycle](https://medium.com/a-journey-with-go/go-timers-life-cycle-403f3580093a)