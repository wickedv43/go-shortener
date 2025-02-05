package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/wickedv43/go-shortener/internal/storage"
)

var errConflict = errors.New("conflict")

type requestJSON struct {
	URL string `json:"url"`
}

type responseJSON struct {
	Result string `json:"response"`
}

func (s *Server) create(c echo.Context) error {
	if c.Request().Header.Get("Content-Type") == "application/json" {
		return c.JSON(http.StatusBadRequest, "Bad request")
	}

	body := c.Request().Body

	url, err := io.ReadAll(body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	data, err := s.save(string(url))
	resURL := fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, data.ShortURL)
	s.logger.Infof("Creating new URL: %s err %s", url, err)

	if err != nil {
		if errors.Is(err, errConflict) {
			c.Response().WriteHeader(http.StatusConflict)
			_, err = c.Response().Write([]byte(resURL))
			return err
		}

		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	c.Response().Header().Set("Content-Type", "text/plain")
	c.Response().WriteHeader(http.StatusCreated)

	_, err = c.Response().Write([]byte(resURL))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	return nil
}

func (s *Server) getShort(c echo.Context) error {
	s.logger.Infof("Getting URL: %s", c.Request().URL)
	short := c.Param("short")

	data, err := s.get(short)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.Response().Header().Set("Location", data.OriginalURL)
	c.Response().WriteHeader(http.StatusTemporaryRedirect)
	return nil
}

func (s *Server) createJSON(c echo.Context) error {
	var (
		url requestJSON
		res responseJSON
	)

	err := c.Bind(&url)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	s.logger.Infof("Creating new URL: %s", url)

	data, err := s.save(url.URL)

	res.Result = fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, data.ShortURL)

	if err != nil {
		if errors.Is(err, errConflict) {
			return c.JSON(http.StatusConflict, res)
		}

		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	s.logger.Infof("Created new URL: %s", res)
	return c.JSON(http.StatusCreated, res)
}

func (s *Server) ping(c echo.Context) error {
	err := s.storage.HealthCheck()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, nil)
}

func (s *Server) batch(c echo.Context) error {
	type batchRequest struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	type batchResponse struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	var (
		reqs []batchRequest
		resp = make([]batchResponse, 0)
		err  error
		data storage.Data
	)

	err = c.Bind(&reqs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if len(reqs) == 0 {
		return c.JSON(http.StatusBadRequest, gin.H{"error": "empty requestJSON"})
	}

	for _, req := range reqs {
		data, err = s.save(req.OriginalURL)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		res := fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, data.ShortURL)

		r := batchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      res,
		}
		resp = append(resp, r)
	}

	return c.JSON(http.StatusCreated, resp)
}
