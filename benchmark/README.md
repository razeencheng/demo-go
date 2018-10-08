# [Go学习笔记(三) | 怎么写Go基准测试（性能测试）](https://razeen.me/post/go-how-to-write-benchmark.html)



或许你经常会思考这样的问题，我用不同的方法实现了同样的效果，哪个会更快？哪个内存消耗更小？这时候你一个简单的基准测试就能解决你的疑惑。


<!-- more -->


Go向来是以工具丰富而著称的，在学习Go的过程中，你会发现无论是写一个单元测试，还是做一些竞争检测都能很快的上手，而且用的很痛快。当然，接下来要说的基准测试也一样。

基准测试工具就在Go的测试包中，下面就用[一个例子](https://github.com/razeencheng/demo-go/tree/master/benchmark)来介绍。



### 举个栗子



由于一些场景需要，我需要将`[]byte`输出16进制字符。

有时候我会这么写:

``` golang
fmt.Sprintf("%x", b)
```

但有时候我会这么写：

```
hex.EncodeToString(b)
```

但到底哪种写法更好呢？今天我就来比较一下。

直接写了个`main.go`

```Golang
func EncodeA(b []byte) string {
	return fmt.Sprintf("%x", b)
}

func EncodeB(b []byte) string {
	return hex.EncodeToString(b)
}
```

再写个测试`main_test.go`

```golang
var buf = []byte("skdjadialsdgasadasdhsakdjsahlskdjagloqweiqwo")

func BenchmarkEncodeA(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeA(buf)
	}
}

func BenchmarkEncodeB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeB(buf)
	}
}
```



就这么简单，我们的基本测试就写完了。从我的写法中你也许就知道：

- 和单元测试一样，都写在`_test.go`文件中；
- 需要以`Benchmark`为函数名开头；
- 和单元测试类似，必须接受一个`*testing.B`参数；
- 被测试代码放在一个循环中。



我们直接跑一下。当然我们也是用`go test`来执行测试，简单的测试只要带上`-bench=.`就可以了。

```Shell
$ go test -bench=.
goos: darwin
goarch: amd64
pkg: github.com/razeencheng/demo-go/benchmark
BenchmarkEncodeA-8       5000000               265 ns/op
BenchmarkEncodeB-8      10000000               161 ns/op
PASS
ok      github.com/razeencheng/demo-go/benchmark        3.397s
```

前两行是平台信息，第三行包名。第四、五行就是测试的结果了。

- `BenchmarkEncodeA-8 ` ,`BenchmarkEncodeB-8 ` 基准测试函数名-GOMAXPROCS
- `5000000`,`10000000` 被测试的函数执行次数，也就是`EncodeA()`被执行了5000000次，`EncodeB()`被执行了10000000次，也就是`b.N`的值了。
- `265 ns/op`,`161 ns/op`表示每次调用被测试函数花费的时间。

从花费的时间上来看，我们知道`EncodeB()`要快一点。



### 更多

你以为就这么简单的结束了么？NONONO。

- `-bench` 可接收一个有效的正则表达式来执行符合条件的测试函数。当你的函数很多时，可以用它来过滤.

```Shell
$ go test -bench=BenchmarkEncodeA
goos: darwin
goarch: amd64
pkg: github.com/razeencheng/demo-go/benchmark
BenchmarkEncodeA-8       5000000               256 ns/op
PASS
ok      github.com/razeencheng/demo-go/benchmark        1.575s
```

- ` -benchmem`可以查看内存分配

``` Shell
$ go test -bench=. -benchmem
goos: darwin
goarch: amd64
pkg: github.com/razeencheng/demo-go/benchmark
BenchmarkEncodeA-8       5000000               261 ns/op             128 B/op          2 allocs/op
BenchmarkEncodeB-8      10000000               162 ns/op             192 B/op          2 allocs/op
PASS
ok      github.com/razeencheng/demo-go/benchmark        3.408s
```

其中`B/op` 表示每次执行会分配多少内存，`allocs/op`表示每次执行会发生多少次内存分配。

- `-benchtime`指定每个测试执行的时间。默认`1s`,当你的函数比较耗时你可以设置更长一点。因为`b.N`是与这个时间有关的。

  当你的运行时间没达到`-benchtime`制定的时间前，`b.N`将以1，2，5，10，20，50…增加，然后重新运行测试代码。

  ```Shell
  $ go test -bench=. -benchmem -benchtime=5s
  goos: darwin
  goarch: amd64
  pkg: github.com/razeencheng/demo-go/benchmark
  BenchmarkEncodeA-8      30000000               254 ns/op             128 B/op          2 allocs/op
  BenchmarkEncodeB-8      50000000               160 ns/op             192 B/op          2 allocs/op
  PASS
  ok      github.com/razeencheng/demo-go/benchmark        16.113s  
  ```

- `-count`指定每个测试执行的次数。

```Shell
$ go test -bench=. -benchmem -count=3
goos: darwin
goarch: amd64
pkg: github.com/razeencheng/demo-go/benchmark
BenchmarkEncodeA-8       5000000               256 ns/op             128 B/op          2 allocs/op
BenchmarkEncodeA-8       5000000               255 ns/op             128 B/op          2 allocs/op
BenchmarkEncodeA-8       5000000               253 ns/op             128 B/op          2 allocs/op
BenchmarkEncodeB-8      10000000               163 ns/op             192 B/op          2 allocs/op
BenchmarkEncodeB-8      10000000               160 ns/op             192 B/op          2 allocs/op
BenchmarkEncodeB-8      10000000               160 ns/op             192 B/op          2 allocs/op
PASS
ok      github.com/razeencheng/demo-go/benchmark        9.984s
```

我常用的也就这些了。

但对于`testing.B`来说，它拥有了`testing.T`的全部接口，所以`Fail,Skip,Error`这些都可以用，而且还增加了

- `SetBytes( i uint64)` 统计内存消耗。
- `SetParallelism(p int)` 制定并行数目。
- `StartTimer / StopTimer / ResertTimer` 操作计时器。

你可以按需使用。



### 注意

`b.N`为一个自增字段，谨慎用它做函数参数。
