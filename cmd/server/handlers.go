package server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

type Expand struct {
	URL string `json:"url"`
}

type Result struct {
	Result string `json:"result"`
}

func (s *Server) addNew(c *gin.Context) {
	url, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	ok, short := s.storage.InStorage(string(url))
	if !ok {
		short = Shorting()
		s.storage.Put(string(url), short)
	}

	c.Header("Content-Type", "text/plain")

	resURL := fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, short)

	c.Writer.WriteHeader(http.StatusCreated)
	_, err = c.Writer.Write([]byte(resURL))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

}

func (s *Server) getShort(c *gin.Context) {
	short := c.Param("short")

	respURL, ok := s.storage.Get(short)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "short not found"})
	}

	c.Header("Location", respURL)
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)
}

func (s *Server) addNewJSON(c *gin.Context) {
	var url Expand
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	err = json.Unmarshal(body, &url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	ok, short := s.storage.InStorage(url.URL)
	if !ok {
		short = Shorting()
		s.storage.Put(url.URL, short)
	}

	var res Result
	res.Result = fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, short)

	c.JSON(http.StatusCreated, res)
}
