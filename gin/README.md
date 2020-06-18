# [gin文件上传与下载]()

Gin是用Go编写的web框架。性能还不错，而且使用比较简单，还支持RESTful API。

日常的使用中我们可能要处理一些文件的上传与下载，我这里简单总结一下。


### 单文件上传

我们使用`multipart/form-data`格式上传文件，利用`c.Request.FormFile`解析文件。

``` golang
// HandleUploadFile 上传单个文件
func HandleUploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "文件上传失败"})
		return
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "文件读取失败"})
		return
	}

	fmt.Println(header.Filename)
	fmt.Println(string(content))
	c.JSON(http.StatusOK, gin.H{"msg": "上传成功"})
}
```

我们上传文件可以看到。

![jietu20180906-002227](https://st.razeen.me/bcj/201809/jietu20180906-002227.png)

我们已经看到文件上传成功，已经文件名字与内容。


### 多文件上传

多文件的上传利用`c.Request.MultipartForm`解析。

``` golang
// HandleUploadMutiFile 上传多个文件
func HandleUploadMutiFile(c *gin.Context) {
	// 限制放入内存的文件大小
	err := c.Request.ParseMultipartForm(4 << 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "文件太大"})
		return
	}
	formdata := c.Request.MultipartForm
	files := formdata.File["file"]

	for _, v := range files {
		file, err := v.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "文件读取失败"})
			return
		}
		defer file.Close()

		content, err := ioutil.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "文件读取失败"})
			return
		}

		fmt.Println(v.Filename)
		fmt.Println(string(content))
	}

	c.JSON(http.StatusOK, gin.H{"msg": "上传成功"})
}
```

多个文件，遍历文件内容即可读取。

~~利用`c.Request.ParseMultipartForm()`可设置上传文件的大小，这里限制了4MB。~~
 `c.Request.ParseMultipartForm()`并不能限制上传文件的大小，只是限制了上传的文件读取到内存部分的大小，如果超过了就存入了系统的临时文件中。
如果需要限制文件大小，需要使用`github.com/gin-contrib/size`中间件，如demo中使用`r.Use(limits.RequestSizeLimiter(4 << 20))`限制最大4Mb。

我们看到

![jietu20180906-002143](https://st.razeen.me/bcj/201809/jietu20180906-002143.png)

两个文件已经上传成功。


### 文件下载

文件的下载主要是注意设置文件名，文件类型等。

``` golang
// HandleDownloadFile 下载文件
func HandleDownloadFile(c *gin.Context) {
	content := c.Query("content")

	content = "hello world, 我是一个文件，" + content

	c.Writer.WriteHeader(http.StatusOK)
	c.Header("Content-Disposition", "attachment; filename=hello.txt")
	c.Header("Content-Type", "application/text/plain")
	c.Header("Accept-Length", fmt.Sprintf("%d", len(content)))
	c.Writer.Write([]byte(content))
}
```

通过
- `Content-Disposition`设置文件名字；
- `Content-Type`设置文件类型，可以到[这里](http://www.runoob.com/http/http-content-type.html)查阅；
- `Accept-Length`这个设置文件长度；
- `c.Writer.Write`写出文件。

成功下载可以看到：

![jietu20180906-004014](https://st.razeen.me/bcj/201809/jietu20180906-004014.png)


* 完整demo[在这里](https://github.com/razeencheng/demo-go/blob/master/gin/gin.go)
