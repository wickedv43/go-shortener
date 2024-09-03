package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/cmd/config"
	"github.com/wickedv43/go-shortener/internal/logger"
)

func main() {
	config.ParseFlags()

	r := gin.New()
	r.Use(gin.Recovery(), logger.Logger())

	S.Init()

	r.POST(`/`, addNew)
	r.GET(`/:short`, getShort)

	err := r.Run(config.FlagRunAddr)
	if err != nil {
		logrus.Fatal(err)
	}
}
