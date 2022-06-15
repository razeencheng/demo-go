//go:build doc
// +build doc

package main

import (
	_ "github.com/razeencheng/demo-go/swaggo-gin/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	swagHandler = ginSwagger.WrapHandler(swaggerfiles.Handler)
}
