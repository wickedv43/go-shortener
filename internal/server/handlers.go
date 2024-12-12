package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/wickedv43/go-shortener/internal/storage"
)

var errConflict = errors.New("conflict")

type expand struct {
	URL string `json:"url"`
}

type result struct {
	Result string `json:"result"`
}

func (s *Server) addNew(c *gin.Context) {
	if c.Request.Header.Get("Content-Type") == "application/json" {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	url, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain")

	short, err := s.save(string(url))
	resURL := fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, short)
	if err != nil {
		if errors.Is(err, errConflict) {
			c.Writer.WriteHeader(http.StatusConflict)
			c.Writer.Write([]byte(resURL))
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Writer.WriteHeader(http.StatusCreated)

	_, err = c.Writer.Write([]byte(resURL))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func (s *Server) getShort(c *gin.Context) {
	short := c.Param("short")

	respURL, ok := s.storage.Get(short)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "short not found"})
		return
	}

	if respURL == "" {
		s.logger.WithField("get", short).Warn("Empty URL returned for short")
	} else {
		s.logger.WithField("get", short).Infoln("Redirecting to URL:", respURL)
	}

	s.logger.WithField("get", short).Infoln("Redirecting to:", respURL)

	c.Redirect(http.StatusTemporaryRedirect, respURL)
}

func (s *Server) addNewJSON(c *gin.Context) {
	var (
		url expand
		res result
	)

	err := c.BindJSON(&url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	short, err := s.save(url.URL)
	res.Result = fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, short)
	if err != nil {
		if errors.Is(err, errConflict) {
			c.JSON(http.StatusConflict, res)
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, res)
}

func (s *Server) ping(c *gin.Context) {
	err := s.storage.Ping()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, nil)
}

func (s *Server) batch(c *gin.Context) {
	type batchRequest struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	type batchResponse struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	var (
		reqs  []batchRequest
		resp  = make([]batchResponse, 0)
		err   error
		short string
	)

	err = c.BindJSON(&reqs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(reqs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty request"})
		return
	}

	for _, req := range reqs {
		short, err = s.save(req.OriginalURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		res := fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, short)

		r := batchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      res,
		}
		resp = append(resp, r)
	}

	c.JSON(http.StatusCreated, resp)
}

func (s *Server) urls(c *gin.Context) {
	type response struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}

	res := make([]response, 0)

	for _, d := range s.storage.DB {
		res = append(res, response{
			ShortURL:    d.ShortURL,
			OriginalURL: d.OriginalURL,
		})
	}

	if len(res) == 0 {
		c.JSON(http.StatusNoContent, nil)
	}

	c.JSON(http.StatusOK, res)
}

func (s *Server) save(url string) (string, error) {
	short, ok := s.storage.InStorage(url)
	if ok {
		return short, errConflict
	} else {
		short = Shorting()
		d := storage.Data{
			OriginalURL: url,
			ShortURL:    short,
		}

		err := s.storage.Put(d)
		if err != nil {
			s.logger.WithError(err).Error("Failed to save data to storage")
			return "", err
		}

		return short, nil
	}
}
