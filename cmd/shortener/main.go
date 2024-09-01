package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	r := gin.Default()
	S.Init()

	r.POST(`/`, addNew)
	r.GET(`/:short`, getShort)

	err := r.Run("localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
}
