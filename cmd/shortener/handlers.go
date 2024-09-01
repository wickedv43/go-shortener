package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
)

func addNew(c *gin.Context) {
	url, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ok, short := S.InStorage(string(url))
	if !ok {
		short = Shorting()
		S.Save(string(url), short)
	}

	log.Println(string(url), short)

	c.Header("Content-Type", "text/plain")

	resURL := "http://localhost:8080/" + short

	c.Writer.WriteHeader(http.StatusCreated)
	_, err = c.Writer.Write([]byte(resURL))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}

func getShort(c *gin.Context) {
	short := c.Param("short")

	respURL, ok := S.Get(short)
	fmt.Println(respURL, short)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "short not found"})
	}

	c.Header("Location", respURL)
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}
