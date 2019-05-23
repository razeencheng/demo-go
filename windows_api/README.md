[## 如何用Go调用Windows API](https://razeencheng.com/post/breaking-all-the-rules-using-go-to-call-windows-api.html)


有时候为了更好的兼容Windows, 或者我们为了获得更高级别功能的访问权限（如配置或创建JobObjects或安全令牌等），我们需要直接去调用Windows的系统API。 很幸运，我们可以利用`syscall`包与系统直接通信，不用用到`CGO` 。 然而，也有不方便的地方，如大多数的API，我们需要依赖不安全 `(unsafe)`的内存管理。



这篇文章，主要记录了我在平时开发过程中以及网上收集到的一些关于Windows API调用相关的知识，或者开发模式，方便你遇到类似的情况后，能更快入手。



<!--more-->

> 注 <sup>1</sup> 本文完整Demo在[这里](https://github.com/razeencheng/demo-go/tree/master/windows_api)。
> 注<sup>2</sup> 文章中并未严格区分过程与函数。



###  关于`syscall` 包

在Go中，`syscall`包会由于你指定的系统或架构的不同而编译出不同的结果，因为`syscall`包里需要编译的函数或类型会根据你指定的编译参数不同而不同。在导入`syscall`时你必须在代码中指定"build tags" 或 用指定的文件后缀来命名你的文件。 Dave Cheney [有篇文章](<https://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool>)深入的介绍了`go build`机制，可以看一看。简单来说，

- 如果你的文件命名结构是这样的，`name_{GOOS}_{GOARCH}.go` 或者 `name_{GOOS}.go`, 那么这个文件只有在指定的`GOOS`+指定的`GOARCH`上才会编译。如：`myfile_windows_amd64.go`只会在`amd64`架构CPU的Windows上才会编译。 而`myfile_windows.go`会在Windows上编译，就不限制CPU架构了。
- 如果你在go代码顶部增加`// +build windows,amd64` 注释，那么该文件只会在`amd64`架构CPU的Windows上才会编译。



### 关于 `unsafe`包

下面是一段`Youtube`视频(需要代理)。

<iframe width="560" height="315" src="https://www.youtube.com/embed/PAAkCSZUG1c?start=830" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

视频中 `Rob Pike`提到：

*With the unsafe package there are no guarantees.*

是的，`Rob Pike`不推荐使用`unsafe`包，因为它**没有任何保障**。



那么，为什么说使用`unsafe`包没有任何保障呢？

1. Go (运行时) 不能保证内置类型（如切片、字符串）在不同的Go版本中内存结构是完全一样的。而且作为支持垃圾回收的语言，开发者并不知道Go内存管理的细节。`unsafe`包会暴露一些内部实现或实际的内存地址，这可能会让你做一些超出预期的事情，如你不小心改变了某个指针指向的地址。

2. Go (语言层面) 不能保证不同版本之间会有相同的特征或者函数签名，换句话来说，就是在 [Go 1.x的兼容性承诺](https://golang.org/doc/go1compat)中,并不包含`unsafe`包。

   

> **Warning:** *Avoid `unsafe` like the plague; if you can help it.*



这两点都告诉我们，在使用`unsafe`包的时候，我们需要特别的注意应该怎么去使用。我们必须了解，用`unsafe`包操作内存时，我们能做什么和不能做什么。而且这也可能会因为不同的Go版本而发生变化，在`unsafe`的[官方文档](https://golang.org/pkg/unsafe/)中，我们能了解到哪些我们该做，哪些不该做，我们应该密切关注。



**Note:**  从技术上来说，`syscall`包，也不在 [Go 1.x的兼容性承诺](https://golang.org/doc/go1compat)中，因为它也不能保证系统是否向后兼容。不过，从Go1.4开始, go底层基本稳定，只有在操作系统发生变化才会有可能发生更改。而调用`Windows DLL`的部分改变的可能性比较小，这点对我们是个好消息。



在[`x/sys/windows`](https://godoc.org/golang.org/x/sys/windows)包中，包含了Go1.x中使用的所有的Windows API调用，你可以直接拿来使用，但注意以下几点：

1. 该包不在[Go 1.x的兼容性承诺](https://golang.org/doc/go1compat)中, 不能保证你的代码稳定，如果想保持稳定，可以切到稳定的Git版本中。
2. 该包的目标也不是暴露所有的Windows API, 而是为Go标准库其他包提供更便携的接口，如`os`,`time`和`net`包。所以你需要的内容，不一定能在该包找到。



虽然是这样，但是现在我们知道了该用那些包去调用Windows API了，同时我们也要知道这会有一定的风险。



### Windows API

Microsoft 提供了大部分的[Windows API](https://docs.microsoft.com/en-us/windows/desktop/apiindex/api-index-portal)文档。API是通过Windows安装时的[DLL(Dynamic Link Library)]([https://zh.wikipedia.org/wiki/%E5%8A%A8%E6%80%81%E9%93%BE%E6%8E%A5%E5%BA%93](https://zh.wikipedia.org/wiki/动态链接库))发布的。DLL是否可用取决于Windows的版本，但API文档中都会列出API什么时候启用，过时或废弃。



### 加载DLL

要在Go中加载DLL，可以使用`syscall.NewLazyDLL`或`syscall.LoadLibrary` 。

- `NewLazyDLL`返回一个`*LazyDLL`，懒加载，只在第一次调用其函数时才加载库; 

- `LoadLibrary`是立即加载DLL库。

其实在`golang.org/x/sys/windows`还支持`windows.NewLazySystemDLL`的方式加载。这是一种安全的加载方式，它能确保DLL搜索路径被绑定到了Windows系统目录。



### 创建函数

当我们加载（懒加载）了DLL库过，我们就要使用`dll.NewProc("ProcName")`去引用一些DLL中的函数(过程)。如：

```go
var（
    kernel32DLL = syscall.NewLazyDLL（“kernel32.dll”）
    procOpenProcess = kernel32DLL.NewProc（“OpenProcess”）
）
```

一旦有个这些引用，我们就可以`Call`这个函数本身的方法，或者使用`syscall.Syscall`函数及其变体进行API调用。使用的过程中发现`Call`方法更方便，但`syscall.Syscall`性能更优。根据函数参数的多数，我们可以使用

`syscall.Syscall`的变体。

- `syscall.Syscall` ：少于4个参数
- `syscall.Syscall6`：4到6个参数
- `syscall.Syscall9`：7到9个参数
- `syscall.Syscall12`：10到12个参数
- `syscall.Syscall15`：13到15个参数

目前Go v1.12中，无法调用超过15个参数的函数。虽然我从来没有遇到过，但在[于OpenGL中](https://github.com/golang/go/issues/28434)确实有这种情况。



### API函数签名

在实际调用DLL函数之前，我们必须要了解一下过程所需要的参数，类型，大小。Microsoft将此描述为Windows API文档的一部分。如`CreateJobObjectA`的过程签名如下：

```c++
HANDLE CreateJobObjectA(
  LPSECURITY_ATTRIBUTES lpJobAttributes,
  LPCSTR                lpName
);
```

也就是说，`CreateJobObjectA`需要一个指向`LPSECURITY_ATTRIBUTES`结构的指针，和一个指向C-String的指针（ASCII编码，技术上是[Windows-1252编码](https://en.wikipedia.org/wiki/Windows-1252) ;它与ASCII兼容）。



### C结构与Go结构

在文档中我们可以搜索到，`LPSECURITY_ATTRIBUTES`是这么定义的：

``` c++
typedef struct _SECURITY_ATTRIBUTES {
  DWORD  nLength;
  LPVOID lpSecurityDescriptor;
  BOOL   bInheritHandle;
} SECURITY_ATTRIBUTES, *PSECURITY_ATTRIBUTES, *LPSECURITY_ATTRIBUTES;
```

这时，我们就必须构造一个类似的Go结构来替代它。这时我们可以参考`syscall`中[SecurityAttributes](https://godoc.org/golang.org/x/sys/windows#SecurityAttributes)的定义。

在Windows API中，我们可以看到，`SecurityAttributes`是这么定义的：

``` c++
typedef struct _SECURITY_ATTRIBUTES {
  DWORD  nLength;
  LPVOID lpSecurityDescriptor;
  BOOL   bInheritHandle;
} SECURITY_ATTRIBUTES, *PSECURITY_ATTRIBUTES, *LPSECURITY_ATTRIBUTES;
```

而Go中[SecurityAttributes](https://godoc.org/golang.org/x/sys/windows#SecurityAttributes)的定义为：

```go
type SecurityAttributes struct {
    Length             uint32
    SecurityDescriptor uintptr
    InheritHandle      uint32
}
```

由此我们大概知道， `DWORD`对应Go `uint32`， `LPVOID (* void)`对应`uintptr`，`BOOL`对应`uint32`。所以在你不知道用什么类型来表示C中对应的结构时，你可以去看看`syscall`或`go.sys`库中找找，或许能有收获。Windows一些参考类型[这里](https://docs.microsoft.com/en-us/windows/desktop/WinProg/windows-data-types)也有描述。 



然而，了解下面这些常见C类型与Go类型的对应关系会非常有用。

``` go
type (
	BOOL          uint32
	BOOLEAN       byte
	BYTE          byte
	DWORD         uint32
	DWORD64       uint64
	HANDLE        uintptr
	HLOCAL        uintptr
	LARGE_INTEGER int64
	LONG          int32
	LPVOID        uintptr
	SIZE_T        uintptr
	UINT          uint32
	ULONG_PTR     uintptr
	ULONGLONG     uint64
	WORD          uint16
)
```



### 字符串

在Windows中，一些函数使用的字符串有两种类型：一种是ANSI编码的，一种是UTF-16编码的。

如`CreateProcess`函数。

```c++
var (
    kernel32DLL = syscall.NewLazyDLL("kernel32.dll")
    procCreateProcessA = kernel32DLL.NewProc("CreateProcessA")
    procCreateProcessW = kernel32DLL.NewProc("CreateProcessW")
)
```

不管是哪一种，我们都不能直接使用Go中的字符串。这就需要我们去做一些兼容。其实这很简单，只要我们在原始字符串后面加上一个零值即可。如下：

``` go
package main

import "unicode/utf16"

// StringToCharPtr converts a Go string into pointer to a null-terminated cstring.
// This assumes the go string is already ANSI encoded.
func StringToCharPtr(str string) *uint8 {
	chars := append([]byte(str), 0) // null terminated
	return &chars[0]
}

// StringToUTF16Ptr converts a Go string into a pointer to a null-terminated UTF-16 wide string.
// This assumes str is of a UTF-8 compatible encoding so that it can be re-encoded as UTF-16.
func StringToUTF16Ptr(str string) *uint16 {
	wchars := utf16.Encode([]rune(str + "\x00"))	
	return &wchars[0]
}
```

其中`StringToUTF16Ptr`在标准库`syscall`中已经有了。



### 调用API

把上面这些知识都用到，我们就可以开始调用一些API了。如我们调用`CreateJobObjectW`。

```go
var (
	kernel32DLL          = syscall.NewLazyDLL("kernel32.dll")
	procCreateJobObjectW = kernel32DLL.NewProc("CreateJobObjectW")
)

// CreateJobObject uses the CreateJobObjectW Windows API Call to create and return a Handle to a new JobObject
func CreateJobObject(attr *syscall.SecurityAttributes, name string) (syscall.Handle, error) {
	r1, _, err := procCreateJobObjectW.Call(
		uintptr(unsafe.Pointer(attr)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))),
	)
	if err != syscall.Errno(0) {
		return 0, err
	}
	return syscall.Handle(r1), nil
}
```

不管调用哪个API，`Call`的模式都是一样的。

而且`syscall.Syscall`函数始终返回`r1,r2 uintptr,err error`， 就最近的实践(windows_amd64)来看，基本可以确定:

- r1 始终返回 `syscall`的值；

- r2 暂且使用；

- err 返回调用Windows API`GetLastError`的结果，这是`syscall`自动调用的。

  

而你传入`Call`中的值必须全部是`uintptr`，不管你原来的类型是什么。但，Go的指针很特别。



由于Go支持垃圾回收，标准的Go指针不是直接指向了物理内存中的一个地址。Go在运行时可以轻松的修改Go指针指向的物理内存地址，如增加堆栈时。当我们把一个Go指针通过`unsafe.Pointer`转换成`uintptr`时，对Go运行时来说，该指针变成了一个未被Go运行时追踪对一个数字而已。即使在下一个指令内，我们也无法确定这个数字指向的是否是它原来指向的那块有效的内存！



正因为如此，我们必须在Syscalls调用时，将指针指向确定的内存。使用`uintptr(unsafe.Pointer(&x))`构造一个参数，告诉编译器，在Syscall期间不能修改x的内存空间。这样，C函数就能正常的去处理该指针了，直到Syscall返回为止。



在[godoc for unsafe.Pointer中](https://golang.org/pkg/unsafe/#Pointer)中写明了四种`unsafe.Pointers`的操作方式原则。这里用到

> (4) Conversion of a Pointer to a uintptr when calling syscall.Syscall.



### 获取原始数据

有时，Syscall会自动为你填充C结构的内存块，如果你要使用就必须将其转化为可用的类型。

许多API的一般调用模式如下：

1. 通过空缓冲区调用一次API，指定一个获取缓冲区长度的变量，获取缓冲区的实际大小；

2. API返回`ERROR_INSUFFICIENT_LENGTH`错误，同时将长度值更新为实际需要的长度；

3. 指定一个实际长度的扩展缓冲区，重新调用；

4. 调用成功。

   

如，我们需要调用`GetExtendedTcpTable`函数。

```c++
IPHLPAPI_DLL_LINKAGE DWORD GetExtendedTcpTable(
  PVOID           pTcpTable,
  PDWORD          pdwSize,
  BOOL            bOrder,
  ULONG           ulAf,
  TCP_TABLE_CLASS TableClass,
  ULONG           Reserved
);
```

`GetExtendedTcpTable`返回的数据为`pTcpTable` 和 `pdwSize` ， 思路如下:

1. 我们第一次将`pTcpTable`直接指一个0值，使用`dwSize`来获取`pTcpTable`的实际长度；
2. 这时，会返回错误`ERROR_INSUFFICIENT_BUFFER`, 同时`dwSize`的值被设置成了`pTcpTable`的实际大小；
3. 指定一个`dwSize`大小的`[]byte`接收数据；
4. 成功。

部分代码如下：

``` go
var (
  iphlpapiDLL             = syscall.NewLazyDLL("iphlpapi.dll")
	procGetExtendedTcpTable = iphlpapiDLL.NewProc("GetExtendedTcpTable")
)

// GetExtendedTcpTable function retrieves a table that contains a list of TCP endpoints available to the application.
func GetExtendedTcpTable(order, ulAf, tableClass uint32) ([]byte, error) {

	var dwSize uint32
	ret, _, err := procGetExtendedTcpTable.Call(
		0,                                // PVOID
		uintptr(unsafe.Pointer(&dwSize)), // PDWORD
		uintptr(order),                   // BOOL
		uintptr(ulAf),                    // ULONG
		uintptr(tableClass),              // TCP_TABLE_CLASS
		0,                                // ULONG
	)
	if ret == 0 {
		return nil, errors.Wrapf(err, "get extended tcp table size failed code %x", ret)
	}

	if syscall.Errno(ret) == syscall.ERROR_INSUFFICIENT_BUFFER {
		buffer := make([]byte, int(dwSize))

		ret, _, err := procGetExtendedTcpTable.Call(
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(unsafe.Pointer(&dwSize)),
			uintptr(order),
			uintptr(ulAf),
			uintptr(tableClass),
			uintptr(uint32(0)),
		)

		if ret != 0 {
			return nil, errors.Wrapf(err, "get extended tcp table failed code %x", ret)
		}

		return buffer, nil
	}

	return nil, errors.Wrapf(err, "get extended tcp table size failed code %x", ret)
}
```



如果你看过[上面函数的API](<https://docs.microsoft.com/en-us/windows/desktop/api/iphlpapi/nf-iphlpapi-getextendedtcptable>)，你应该会知道输入参数`ulAf`和`TableClass`的值 决定了输出的buffer具体的内容。

如果我们输入的是`AF_INET + TCP_TABLE_OWNER_PID_ALL` 那么我们得到的数据的实际结构应该是[`MIB_TCPTABLE_OWNER_PID`](<https://docs.microsoft.com/en-us/windows/desktop/api/tcpmib/ns-tcpmib-_mib_tcptable_owner_pid>)，其结构如下：

```c++
typedef struct _MIB_TCPTABLE_OWNER_PID {
  DWORD                dwNumEntries;
  MIB_TCPROW_OWNER_PID table[ANY_SIZE];
} MIB_TCPTABLE_OWNER_PID, *PMIB_TCPTABLE_OWNER_PID;
```

这里第一个参数`dwNumEntries`指明有`MIB_TCPROW_OWNER_PID` table的数量。

而第二个参数则是一个变长的数组。。。 那么我们该怎么用Go去表示呢？



### 处理变长数据

其实我们可以利用数组来创建一个兼容该结构的Go结构，这要得益于Go中数组的内存布局为连续的内存区域。

我们定义的对应结构如下：

```go
type MIB_TCPTABLE_OWNER_PID struct {
	dwNumEntries uint32
	table        [1]MIB_TCPROW_OWNER_PID
}

type MIB_TCPROW_OWNER_PID struct {
	dwState      uint32
	dwLocalAddr  [4]byte
	dwLocalPort  uint32
	dwRemoteAddr [4]byte
	dwRemotePort uint32
	dwOwningPid  uint32
}
```

你会说，怎么`table`的长度只为1，这里先存个疑问。



现在我们首先要知道`dwNumEntries`的大小，我们才能确定table的数量。于是利用`unsafe.Pointer`将buffer内的数据转换为Go结构。

```go
	pTable := (*MIB_TCPTABLE_OWNER_PID)(unsafe.Pointer(&buffer[0]))
```

这里，我们将一个指针指向缓冲区的第一个字节的内存地址，然后利用`unsafe.Pointer`我们可以将该指针转换为任意类型的指针。其实这个操作是非常危险的，如果你不知道为什么要转换的话。这里我们能转化是因为我们遵循`unsafe.Pointer`[文档中的第一条](https://golang.org/pkg/unsafe/)：

> (1) Conversion of a *T1 to Pointer to *T2.
>
> Provided that T2 is no larger than T1 and that the two share an equivalent memory layout, this conversion allows reinterpreting data of one type as data of another type. 

也就是说，要把\*T1转换为\*T2，那么T2的长度必须比T1的长，而且T1,T2的内存布局必须相同。



那么，刚刚我们把table 的类型指定为`[1]MIB_TCPROW_OWNER_PID`, 长度为1显然不是正确的大小。但这没关系，因为`[1]MIB_TCPROW_OWNER_PID`的长度肯定是小于实际`[1+N]MIB_TCPROW_OWNER_PID`的长度的。而且他们的内存布局是一样的。

由于此时，我们已经知道了`dwNumEntries`的大小，我们可以使用`unsafe.Pointer`的另一规则来遍历获取数组。

> (3) Conversion of a Pointer to a uintptr and back, with arithmetic.

```go
	rows := make([]MIB_TCPROW_OWNER_PID, int(pTable.dwNumEntries))
	for i := 0; i < int(pTable.dwNumEntries); i++ {
		rows[i] = *(*MIB_TCPROW_OWNER_PID)(unsafe.Pointer(
			uintptr(unsafe.Pointer(&pTable.table[0])) +
				uintptr(i)*unsafe.Sizeof(pTable.table[0])))
	}
```

在这里，我们利用`规则 (3)`迭代已知长度的数组，因为我们知道第一个元素的位置，每个元素的大小，元素的个数，以及结构在内存中的布局是连续的。



这里还有一个更简单的方法，能让我们直接获取table的数据:

```go
	rows2 := ((*[1 << 30]MIB_TCPROW_OWNER_PID)(unsafe.Pointer(&pTable.table[0]))[:int(pTable.dwNumEntries):int(pTable.dwNumEntries)])
```

这种做法一开始就将这个指针转换成一个非常大的数组指针，然后使用正确的长度取获取实际的内容。好处是不用创建其他切片，复制数据；缺点就是我们需要分配一个足够大的内存去接收，这个大小各平台会有一些差异。



你可以在这里体验一下[Go Playground](https://play.golang.org/p/1XN1bLer-se)。



### 最后

现在你应该知道了调用Windows API的一些基本步骤与方法，如果遇到问题可以留言，我们一起解决～



- 原文[Breaking all the rules: Using Go to call Windows API](https://medium.com/jettech/breaking-all-the-rules-using-go-to-call-windows-api-2cbfd8c79724) 有所改动。








