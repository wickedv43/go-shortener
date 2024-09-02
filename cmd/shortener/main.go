package main

import (
	"github.com/gin-gonic/gin"
	"github.com/wickedv43/go-shortener/cmd/config"
	"log"
)

func main() {
	config.ParseFlags()

	r := gin.Default()
	S.Init()

	r.POST(`/`, addNew)
	r.GET(`/:short`, getShort)

	err := r.Run(config.FlagRunAddr)
	if err != nil {
		log.Fatal(err)
	}
}
