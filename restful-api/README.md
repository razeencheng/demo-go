# [Golang中的RESTful API最佳实践](https://razeencheng.com/post/golang-and-restful-api.html)


RESRful  API已经流行很多年了，我也一直在使用它。最佳实践也看过不少，但当一个项目完成，再次回顾/梳理项目时，会发现很多API和规范还是多少有些出入。在这篇文章中，我们结合Go Web再次梳理一下RESTful API的相关最佳实践。



<!--more-->



### 关于RESTful API

关于什么是RESTful API，不再累述。推荐几个相关链接。

- [理解RESTful架构](https://www.ruanyifeng.com/blog/2011/09/restful.html)
- [REST API Tutorial](https://restfulapi.net/)



### 1.使用JSON

不管是接收还是返回数据都推荐使用JSON。

通常返回数据的格式有JSON和XML，但XML过于冗长，可读性差，而且各种语言的解析上也不如JSON，使用JSON的好处，显而易见。

而接收数据，我们这里也推荐使用JSON，对于后端开发而言，入参直接与模型绑定，省去冗长的参数解析能简化不少代码，而且JSON能更简单的传递一些更复杂的结构等。

正如示例代码中的这一段，我们以`gin`框架为例。

```go
// HandleLogin doc
func HandleLogin(c *gin.Context) {
	param := &LoginParams{}
	if err := c.BindJSON(param); err != nil {
		c.JSON(http.StatusBadRequest, &Resp{Error: "parameters error"})
		return
	}

	// 做一些校验
	// ...

	session := sessions.Default(c)
	session.Set(sessionsKey, param.UserID)
	session.Save()
	c.JSON(http.StatusOK, &Resp{Data: "login succeed"})
}
```

通过`c.BindJSON`,轻松的将入参于模型`LoginParams`绑定；通过`c.JSON`轻松的将数据JSON序列化返回。



但所有接口都必须用JSON么？那也未必。比如文件上传，这时我们使用`FormData`比把文件base64之类的放到JSON里面更高效。



### 2.路径中不包含动词

我们的HTTP请求方法中已经有`GET`,`POST`等这些动作了，完全没有必要再路径中加上动词。

我们常用HTTP请求方法包括`GET`,`POST`,`PUT`和`DELETE`, 这也对应了我们经常需要做的数据库操作。`GET`查找/获取资源，`POST`新增资源，`PUT`修改资源，`DELETE`删除资源。

如下，这些路径中没有任何动词，简洁明了。

```go
// 获取文章列表
v1.GET("/articles", HandleGetArticles)
// 发布文章
v1.POST("/articles", HandlePostArticles)
// 修改文章
v1.PUT("/articles", HandleUpdateArticles)
// 删除文章
v1.DELETE("/articles/:id", HandleDeleteArticles)
```



### 3.路径中对应资源用复数

就像我们上面这段代码，`articles`对于的是我们的文章资源，背后就是一张数据库表`articles`, 所以操作这个资源的应该都用复数形式。



### 4.次要资源可分层展示

一个博客系统中，最主要的应该是文章了，而评论应该是其子资源，我们可以评论嵌套在它的父资源后面，如：

``` go
// 获取评论列表
v1.GET("/articles/:articles_id/comments", HandleGetComments)
// 添加评论
v1.POST("/articles/:articles_id/comments", HandleAddComments)
// 修改评论
v1.PUT("/articles/:articles_id/comments/:id", HandleUpdateComments)
// 删除评论
v1.DELETE("/articles/:articles_id/comments/:id", HandleDeleteComments)
```

那么，我们需要获取所有文章的评论怎么办？可以这么写：

``` go
v1.GET("/articles/-/comments", HandleGetComments)
```

但这也不是决对的，资源虽然有层级关系，但这种层级关系不宜太深，个人感觉两层最多了，如果超过，可以直接拿出来放在一级。



### 5.分页、排序、过滤

获取列表时，会使用到分页、排序过滤。一般：

``` sh
?page=1&page_size=10  # 指定页面page与分页大小page_size
?sort=-create_at,+author # 按照创建时间create_at降序，作者author升序排序
?title=helloworld # 按字段title搜索
```



### 6.统一数据格式

不管是路径的格式，还是参数的格式，还是返回值的格式建议统一形式。

一般常用的格式有`蛇形`,`大驼峰`和`小驼峰`，个人比较喜欢`蛇形`。Anyway, 不管哪种，只要统一即可。

除了参数的命名统一外，返回的数据格式，最好统一，方便前端对接。

如下，我们定义`Resp`为通用返回数据结构，`Data`中存放反会的数据，如果出错，将错误信息放在`Error`中。

```go
// Resp doc
type Resp struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

// 登陆成功返回
  c.JSON(http.StatusOK, &Resp{Data: "login succeed"})
// 查询列表
	c.JSON(http.StatusOK, &Resp{Data: map[string]interface{}{
		"result": tempStorage,
		"total":  len(tempStorage),
	}})
// 参数错误
	c.JSON(http.StatusBadRequest, &Resp{Error: "parameters error"})
```



### 7.善用HTTP状态码

HTTP状态码有很多，我们没有必要也不可能全部用上，常用如下：

- 200 StatusOK - 只有成功请求都返回200。
- 400 StatusBadRequest - 当出现参数不对，用户参数校验不通过时，给出该状态，并返回Error
- 401 StatusUnauthorized - 没有登陆/经过认证
- 403 Forbidden - 服务端拒绝授权(如密码错误)，不允许访问
- 404 Not Found - 路径不存在
- 500 Internal Server Error - 所请求的服务器遇到意外的情况并阻止其执行请求
- 502 Bad Gateway - 网关或代理从上游接收到了无效的响应 
- 503 Service Unavailable - 服务器尚未处于可以接受请求的状态

其中`502`,`503`，我们写程序时并不会明确去抛出。所以我们平常用6个状态码已经能很好的展示服务端状态了。

同时，我们将状态与返回值对应起来，`200`状态下，返回`Data`数据；其他状态返回`Error`。



### 8.API版本化

正如Demo中所示，我们将路由分组到了`/api/v1`路径下面，版本化API。如果后续的服务端升级，但可能仍有很大部分客户端请求未升级，依然请求老版本的API，那么我们只需要增加`/api/v2`，然后在该路径下为已升级的客户端提供服务。这样，我们就做到了API的版本控制，可以平滑的从一个版本切换到另外一个版本。

```go
	v1 := r.Group("/api/v1")
	{
		v1.POST("/login", HandleLogin)
		v1.GET("/articles", HandleGetArticles)
		v1.GET("/articles/:id/comments", HandleGetComments)
    // ....
```



### 9. 统一 ‘/‘ 开头

所以路由中，路径都以’/‘开头，虽然框架会为我们做这件事，但还是建议统一加上。



### 10. 增加/更新操作 返回资源

对于`POST`,`PUT`操作，建议操作后，返回更新后的资源。



### 11. 使用HTTPS 

对于暴露出去的接口/OpenAPI，一定使用HTTPS。一般时候，我们可以直接在服务前面架设一个WebServer，在WebServer内部署证书即可。当然，如果是直接由后端暴露出的接口，有必要直接在后端开启HTTPS！



### 12. 规范的API文档

对于我们这种前后端分离的架构，API文档是很重要。在Go中，我们很容易的能用swag结合代码注释自动生成API文档，在[ <使用swaggo自动生成Restful API文档>](https://razeencheng.com/post/go-swagger.html)中，我详细的介绍了怎么生成以及怎么写注释。



### 总结

API写的好不好，重要的还是看是否遵循WEB标准和保持一致性，最终目的也是让这些API更清晰，易懂，安全，希望这些建议对你有所帮助。










