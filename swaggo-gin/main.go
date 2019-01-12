package main

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

// 文档Handle
var swagHandler gin.HandlerFunc

// @title Swagger Example API
// @version 1.0
// @description This is a sample server celler server.
// @termsOfService https://razeen.me

// @contact.name Razeen
// @contact.url https://razeen.me
// @contact.email me@razeen.me

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host 127.0.0.1:8080
// @BasePath /api/v1

func main() {

	r := gin.Default()
	store := sessions.NewCookieStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/hello", HandleHello)
		v1.POST("/login", HandleLogin)
		v1Auth := v1.Use(HandleAuth)
		{
			v1Auth.POST("/upload", HandleUpload)
			v1Auth.GET("/list", HandleList)
			v1Auth.GET("/file/:id", HandleGetFile)
		}
	}

	if swagHandler != nil {
		r.GET("/swagger/*any", swagHandler)
	}

	r.Run(":8080")
}
