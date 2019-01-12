// +build doc

package main

import (
	_ "github.com/razeencheng/demo-go/swaggo-gin/docs"

	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func init() {
	swagHandler = ginSwagger.WrapHandler(swaggerFiles.Handler)
}
