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

var ErrConflict = errors.New("conflict")
var ErrNoContent = errors.New("no content")

type requestJSON struct {
	URL string `json:"url"`
}

type responseJSON struct {
	Result string `json:"result"`
}

func (s *Server) create(c echo.Context) error {
	if c.Request().Header.Get("Content-Type") == "application/json" {
		return c.JSON(http.StatusBadRequest, "Bad request")
	}

	body := c.Request().Body

	s.logger.Infof("Received body: %v", body)
	url, err := io.ReadAll(body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	userID := c.Get("userID").(int)

	data, err := s.save(c.Request().Context(), string(url), userID)
	resURL := fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, data.ShortURL)

	if err != nil {
		if errors.Is(err, ErrConflict) {
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
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	return nil
}

// TODO: add 410 err
func (s *Server) getShort(c echo.Context) error {
	s.logger.Infof("Getting URL: %s", c.Request().URL)
	short := c.Param("short")

	data, err := s.get(c, short)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	if data.DeletedFlag {
		return c.JSON(http.StatusGone, "Gone")
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
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	userID := c.Get("userID").(int)
	s.logger.Infof("User ID: %v", userID)

	data, err := s.save(c.Request().Context(), url.URL, userID)

	res.Result = fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, data.ShortURL)

	if err != nil {
		if errors.Is(err, ErrConflict) {
			return c.JSON(http.StatusConflict, res)
		}

		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

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
		resp []batchResponse
		err  error
		data storage.Data
	)

	userID := c.Get("userID").(int)

	err = c.Bind(&reqs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	if len(reqs) == 0 {
		return c.JSON(http.StatusBadRequest, gin.H{"error": "empty requestJSON"})
	}

	for _, req := range reqs {
		data, err = s.save(c.Request().Context(), req.OriginalURL, userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		res := fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, data.ShortURL)

		r := batchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      res,
		}
		resp = append(resp, r)
		s.logger.Infof("batch response: %v", r)
		s.logger.Infof("batch response: %v", resp)
	}

	s.logger.Infof("batch response: %v", resp)
	return c.JSON(http.StatusCreated, resp)
}

func (s *Server) userURLs(c echo.Context) error {
	userID := c.Get("userID").(int)
	urls, err := s.getAll(c, userID)
	if err != nil {
		if errors.Is(err, ErrNoContent) {
			return c.JSON(http.StatusNoContent, "No content")
		}
		return c.JSON(http.StatusInternalServerError, "getting data")
	}

	for i := range urls {
		urls[i].ShortURL = fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, urls[i].ShortURL)
	}

	return c.JSON(http.StatusOK, urls)
}

func (s *Server) deleteUserURLs(c echo.Context) error {
	var shorts []string

	err := c.Bind(&shorts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "server error")
	}

	//userID check
	okShorts := make([]string, 0)

	userID := c.Get("userID").(int)

	for _, short := range shorts {
		var d storage.Data

		d, err = s.get(c, short)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "server error")
		}

		if !d.DeletedFlag && d.UUID == userID {
			okShorts = append(okShorts, short)
		}
	}

	// send to chan
	for _, short := range okShorts {
		s.urlDeleteChan <- short
	}

	return c.JSON(http.StatusAccepted, nil)

}
