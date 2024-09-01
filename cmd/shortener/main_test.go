package main

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.POST("/", addNew)
	router.GET("/:short", getShort)
	S.Init()
	return router
}
