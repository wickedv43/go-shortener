package server

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/storage"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Expand struct {
	URL string `json:"url"`
}

type Result struct {
	Result string `json:"result"`
}

type batchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type batchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (s *Server) addNew(c *gin.Context) {
	//
	if c.Request.Header.Get("Content-Type") == "application/json" {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	var d storage.Data

	url, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	short, ok := s.storage.InStorage(string(url))
	if !ok {
		short = Shorting()
		d.OriginalURL = string(url)
		d.ShortURL = short
		s.storage.Put(d)
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
	var (
		url Expand
		res Result
	)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	err = json.Unmarshal(body, &url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	short := s.save(url.URL)

	res.Result = fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, short)

	c.JSON(http.StatusCreated, res)
}

func (s *Server) save(url string) string {
	var d storage.Data
	short, ok := s.storage.InStorage(url)
	if !ok {
		short = Shorting()
		d.OriginalURL = url
		d.ShortURL = short
		s.storage.Put(d)
		return short
	}

	s.logger.Logf(logrus.DebugLevel, "saved: %s, short: %s", url, short)

	return short
}

func (s *Server) ping(c *gin.Context) {
	err := s.storage.Ping()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, nil)
}

func (s *Server) batch(c *gin.Context) {
	var (
		reqs []batchRequest
	)

	res := make([]batchResponse, 0)

	err := c.BindJSON(&reqs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(reqs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty request"})
		return
	}

	for _, req := range reqs {
		short := s.save(req.OriginalURL)

		r := batchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      short,
		}
		res = append(res, r)
	}

	c.JSON(http.StatusCreated, res)
}
