package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

const USER = "admin"
const PWD = "admin"

func main() {
	r := gin.Default()

	store := sessions.NewCookieStore([]byte("jdagldagsdadhsbdgaj"))
	store.Options(sessions.Options{
		MaxAge:   7200,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("httpsgateway", store))
	r.NoRoute(func(c *gin.Context) { c.JSON(http.StatusNotFound, "Invaild api request") })

	r.POST("login", HandlleLogin)
	api := r.Group("api", Auth())
	{
		api.GET("logout", HandleLogout)
		api.GET("hello_world", HandleHelloWorld)
	}

	t := r.Group("test")
	{
		t.POST("upload_file", HandleUploadFile)
		t.POST("upload_muti_file", HandleUploadMutiFile)
		t.GET("download", HandleDownloadFile)
	}

	r.RunTLS(":8443", "./cert.pem", "./key.pem")
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		u := session.Get("user")
		if u == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "您暂未登录"})
			c.Abort()
			return
		}
	}
}

func HandlleLogin(c *gin.Context) {
	user := c.PostForm("user")
	password := c.PostForm("password")

	if user != USER || password != PWD {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "用户名或密码不正确"})
		return
	}

	session := sessions.Default(c)
	session.Set("user", USER)
	session.Save()
	c.JSON(http.StatusOK, gin.H{"msg": "login succeed"})

}

func HandleLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("user")
	session.Save()
	c.JSON(http.StatusOK, gin.H{"data": "See you!"})
}

func HandleHelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": "Hello World!"})
}

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

// HandleUploadMutiFile 上传多个文件
func HandleUploadMutiFile(c *gin.Context) {
	// 设置文件大小
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
