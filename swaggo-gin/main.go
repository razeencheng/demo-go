package main

import (
	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	v1 := r.Group("api/v1")
	{
		v1.GET("/hello", HandleHello)
		v1.POST("/login", HandleLogin)
		v1Auth := r.Use(HandleAuth)
		{
			v1Auth.POST("/upload", HandleUpload)
			v1Auth.GET("/list", HandleList)
		}
	}

	r.Run(":8080")
}
