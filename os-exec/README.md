[#Go学习笔记(八) | 使用 os/exec 执行命令](https://razeencheng.com/post/simple-use-go-exec-command.html)

用Go去调用一些外部的命令其实很愉快的，这遍文章就总结一下我自己日常用的比较多的几种方法。



### 关于Unix标准输入输出

在具体聊`os/exec`的使用前，了解一下shell的标准输出是很有必要的。

我们平常会用到或看到这样的命令：

```shell
$ ls xxx 1>out.txt 2>&1
$ nohup xxx 2>&1 &
```

你知道这里`1,2`含义么？



其实这里的`1,2`指的就是Unix文件描述符。文件描述符其实就一数字，每一个文件描述符代表的都是一个文件。如果你打开100个文件，你就会获取到100个文件描述符。



这里需要注意的一点就是，在Unix中[**一切皆文件**](<https://en.wikipedia.org/wiki/Everything_is_a_file>)。当然，这里我们不必去深究，我们需要知道的是`1，2`代表的是标准输出`stdout`与标准错误输出`stderr`。还有`0`代表标准输入`stdin`。



在`os/exec`中就用到了`Stdin`,`Stdout`,`Stderr`，这些基本Unix知识或能帮助我们更好理解这些参数。



### os/exec

`os/exec`包内容并不多，我们大概过一下。

1. [LookPath(file string) (string, error)](https://godoc.org/os/exec#LookPath)

   寻找可执行文件路径，如果你指定的可执行文件在`$PATH`中，就会返回这个可执行文件的相对/绝对路径；如果你指定的是一个文件路径，他就是去判断文件是否可读取/执行，返回的是一样的路径。

   在我们需要使用一些外部命令/可执行文件的时候，我们可以先使用该函数判断一下该命令/可执行文件是否有效。

2. [Command(name string, arg ...string) *Cmd](https://godoc.org/os/exec#Command)

   使用你输入的参数，返回Cmd指针，可用于执行Cmd的方法。

   这里`name`就是我们的命令/可执行文件，后面的参数可以一个一个输入。

3. [CommandContext(ctx context.Context, name string, arg ...string) *Cmd](https://godoc.org/os/exec#CommandContext)

   和上面功能一样，不过我们可以用上下文做一些超时等控制。

4. 之后几个就是Cmd的一些方法。

   - [(c *Cmd) CombinedOutput() ([]byte, error)](https://godoc.org/os/exec#Cmd.CombinedOutput) 将标准输出，错误输出一起返回；

   - [(c *Cmd) Output() ([]byte, error)](https://godoc.org/os/exec#Cmd.Output) 输出标准输出，错误从error返回；

   -  [(c *Cmd) Run() error](https://godoc.org/os/exec#Cmd.Run) 执行任务，等待执行完成；

   - [(c *Cmd) Start() error](https://godoc.org/os/exec#Cmd.Start)， [(c *Cmd) Wait() error](https://godoc.org/os/exec#Cmd.Wait) 前者执行任务，不等待完成，用后者等待，并释放资源；

   - [(c *Cmd) StderrPipe() (io.ReadCloser, error)](https://godoc.org/os/exec#Cmd.StderrPipe) 

      [(c *Cmd) StdinPipe() (io.WriteCloser, error)](https://godoc.org/os/exec#Cmd.StdinPipe)

     [(c *Cmd) StdoutPipe() (io.ReadCloser, error)](https://godoc.org/os/exec#Cmd.StdoutPipe)

     这三个功能类似，就是提供一个标准输入/输出/错误输出的管道，我们可用这些管道中去输入输出。

其实读完，结合官方的一些example，使用很简单，下面具体写几个场景。



*注*

1. 本文全部的Demo在[这里](https://github.com/razeencheng/demo-go/tree/master/os-exec)。

2. `./testcmd/testcmd`是我用Go写的一个简单的可执行文件，可以根据指定的参数 输出/延时输出/输出错误，方便我们演示。如下

```go
func main() {

	var (
		start bool
		e     bool
	)

	flag.BoolVar(&start, "s", false, "start output")
	flag.BoolVar(&e, "e", false, "output err")
	flag.Parse()

	if start {
		for i := 5; i > 0; i-- {
			fmt.Fprintln(os.Stdout, "test cmd output", i)
			time.Sleep(1 * time.Second)
		}
	}

	if e {
		fmt.Fprintln(os.Stderr, "a err occur")
		os.Exit(1)
	}

	fmt.Fprintln(os.Stdout, "test cmd stdout")
}
```



### 简单执行

```go
// 简单执行
func test1() {
	cmd := exec.Command("./testcmd/testcmd", "-s")

	// 使用CombinedOutput 将stdout stderr合并输出
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("test1 failed %s\n", err)
	}
	log.Println("test1 output ", string(out))
}
```

输出：

```shell
Run Test 1
2019/06/06 18:02:39 test1 output  test cmd output 5
test cmd output 4
test cmd output 3
test cmd output 2
test cmd output 1
done
```

整个过程等待5秒，所有结果一次输出。



### 分离标准输出与错误输出

将错误分开输出，同时开了两个协成，同步的接收命令的输出内容。

```go
// 	stdout & stderr 分开输出
func test2() {
	cmd := exec.Command("./testcmd/testcmd", "-s", "-e")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := stderr.Read(buf)

			if n > 0 {
				log.Printf("read err %s", string(buf[:n]))
			}

			if n == 0 {
				break
			}

			if err != nil {
				log.Printf("read err %v", err)
				return
			}
		}
	}()

	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := stdout.Read(buf)

			if n == 0 {
				break
			}

			if n > 0 {
				log.Printf("read out %s", string(buf[:n]))

			}

			if n == 0 {
				break
			}

			if err != nil {
				log.Printf("read out %v", err)
				return
			}

		}
	}()

	err := cmd.Wait()
	if err != nil {
		log.Printf("cmd wait %v", err)
		return
	}
}
```

输出：

```shell
Run Test 2
2019/06/06 18:02:39 read out test cmd output 5
2019/06/06 18:02:40 read out test cmd output 4
2019/06/06 18:02:41 read out test cmd output 3
2019/06/06 18:02:42 read out test cmd output 2
2019/06/06 18:02:43 read out test cmd output 1
2019/06/06 18:02:44 read err a err occur
2019/06/06 18:02:44 cmd wait exit status 1
```



### 按行读取输出内容

使用`bufio`按行读取输出内容。

``` go
// 按行读输出的内容
func test3() {
	cmd := exec.Command("./testcmd/testcmd", "-s", "-e")
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	oReader := bufio.NewReader(stdout)
	eReader := bufio.NewReader(stderr)

	cmd.Start()

	go func() {
		for {
			line, err := oReader.ReadString('\n')

			if line != "" {
				log.Printf("read line %s", line)
			}

			if err != nil || line == "" {
				log.Printf("read line err %v", err)
				return
			}

		}
	}()

	go func() {
		for {
			line, err := eReader.ReadString('\n')

			if line != "" {
				log.Printf("read err %s", line)
			}

			if err != nil || line == "" {
				log.Printf("read err %v", err)
				return
			}

		}
	}()

	err := cmd.Wait()
	if err != nil {
		log.Printf("cmd wait %v", err)
		return
	}
}
```

输出：

``` shell
Run Test 3
2019/06/06 18:06:44 read line test cmd output 5
2019/06/06 18:06:45 read line test cmd output 4
2019/06/06 18:06:46 read line test cmd output 3
2019/06/06 18:06:47 read line test cmd output 2
2019/06/06 18:06:48 read line test cmd output 1
2019/06/06 18:06:49 read err a err occur
2019/06/06 18:06:49 cmd wait exit status 1
```





### 设置执行超时时间

有时候我们要控制命令的执行时间，这是就可以使用上下文去控制了。

``` go
// 通过上下文控制超时
func test4() {

	ctx, calcel := context.WithTimeout(context.Background(), 2*time.Second)
	defer calcel()

	cmd := exec.CommandContext(ctx, "./testcmd/testcmd", "-s", "-e")

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	oReader := bufio.NewReader(stdout)
	eReader := bufio.NewReader(stderr)

	cmd.Start()

	go func() {
		for {
			line, err := oReader.ReadString('\n')

			if line != "" {
				log.Printf("read line %s", line)
			}

			if err != nil || line == "" {
				log.Printf("read line err %v", err)
				return
			}

		}
	}()

	go func() {
		for {
			line, err := eReader.ReadString('\n')

			if line != "" {
				log.Printf("read err %s", line)
			}

			if err != nil || line == "" {
				log.Printf("read err %v", err)
				return
			}

		}
	}()

	err := cmd.Wait()
	if err != nil {
		log.Printf("cmd wait %v", err)
		return
	}
}
```

输出：

```shell
Run Test 4
2019/06/06 18:06:49 read line err EOF
2019/06/06 18:06:49 read err EOF
2019/06/06 18:06:49 read line test cmd output 5
2019/06/06 18:06:50 read line test cmd output 4
2019/06/06 18:06:51 read line err EOF
2019/06/06 18:06:51 read err EOF
2019/06/06 18:06:51 cmd wait signal: killed
```



### 持续输入指令，交互模式

有很多命令支持交互模式，我们进入之后就可以持续的输入一些命令，同时获取输出。如`openssl`命令。

下面我们需要进入交换模式，执行输入三个命令，并获取输出。

``` go
// 持续输入
func test5() {
	cmd := exec.Command("openssl")

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	stdin, _ := cmd.StdinPipe()

	cmd.Start()

	// 读
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := stderr.Read(buf)

			if n > 0 {
				fmt.Println(string(buf[:n]))
			}

			if n == 0 {
				break
			}

			if err != nil {
				log.Printf("read err %v", err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := stdout.Read(buf)

			if n == 0 {
				break
			}

			if n > 0 {
				fmt.Println(string(buf[:n]))
			}

			if n == 0 {
				break
			}

			if err != nil {
				log.Printf("read out %v", err)
				return
			}

		}
	}()

	// 写
	go func() {
		stdin.Write([]byte("version\n\n"))
		stdin.Write([]byte("ciphers -v\n\n"))
		stdin.Write([]byte("s_client -connect razeencheng.com:443"))
		stdin.Close()
		wg.Done()
	}()

	wg.Wait()
	err := cmd.Wait()
	if err != nil {
		log.Printf("cmd wait %v", err)
		return
	}

}
```

这里，我们就用到了`stdin`标准输入了。输出如下：

``` shell
Run Test 5
OpenSSL> LibreSSL 2.6.5
OpenSSL> OpenSSL>
ECDHE-RSA-AES256-GCM-SHA384 TLSv1.2 Kx=ECDH     Au=RSA  Enc=AESGCM(256) Mac=AEAD
ECDHE-ECDSA-AES256-GCM-SHA384 TLSv1.2 Kx=ECDH     Au=ECDSA Enc=AESGCM(256) Mac=AEAD
ECDHE-RSA-AES256-SHA384 TLSv1.2 Kx=ECDH     Au=RSA  Enc=AES(256)  Mac=SHA384
ECDHE-ECDSA-AES256-SHA384 TLSv1.2 Kx=ECDH     Au=ECDSA Enc=AES(256)  Mac=SHA384
ECDHE-RSA-AES256-SHA    SSLv3 Kx=ECDH     Au=RSA  Enc=AES(256)  Mac=SHA1
ECDHE-ECDSA-AES256-SHA  SSLv3 Kx=ECDH     Au=ECDSA Enc=AES(256)  Mac=SHA1
DHE-RSA-AES256-GCM-SHA384 TLSv1.2 Kx=DH       Au=RSA  Enc=AESGCM(256) Mac=AEAD
DHE-RSA-AES256-SHA256   TLSv1.2 Kx=DH       Au=RSA  Enc=AES(256)  Mac=SHA256
DES-CBC-SHA             SSLv3 Kx=RSA      Au=RSA  Enc=DES(56)   Mac=SHA1
...
OpenSSL> OpenSSL>
4466583148:error:14004410:SSL routines:CONNECT_CR_SRVR_HELLO:sslv3 alert handshake failure:/BuildRoot/Library/Caches/com.apple.xbs/Sources/libressl/libressl-22.260.1/libressl-2.6/ssl/ssl_pkt.c:1205:SSL alert number 40
4466583148:error:140040E5:SSL routines:CONNECT_CR_SRVR_HELLO:ssl handshake failure:/BuildRoot/Library/Caches/com.apple.xbs/Sources/libressl/libressl-22.260.1/libressl-2.6/ssl/ssl_pkt.c:585:

CONNECTED(00000005)
---
no peer certificate available
---
No client certificate CA names sent
---
SSL handshake has read 7 bytes and written 0 bytes
---
New, (NONE), Cipher is (NONE)
Secure Renegotiation IS NOT supported
Compression: NONE
Expansion: NONE
No ALPN negotiated
SSL-Session:
    Protocol  : TLSv1.2
    Cipher    : 0000
    Session-ID:
    Session-ID-ctx:
    Master-Key:
    Start Time: 1559815613
    Timeout   : 7200 (sec)
    Verify return code: 0 (ok)
---

```





—END---