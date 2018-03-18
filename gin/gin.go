package main

import (
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

	r.RunTLS(":443", "./cert.pem", "./key.pem")
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
