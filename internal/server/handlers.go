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

// ErrConflict is returned when a short URL already exists.
var ErrConflict = errors.New("conflict")

// ErrNoContent is returned when no data is available for the user.
var ErrNoContent = errors.New("no content")

// requestJSON represents the structure of incoming JSON with a URL.
type requestJSON struct {
	URL string `json:"url"`
}

// responseJSON represents the structure of the response containing the short URL.
type responseJSON struct {
	Result string `json:"result"`
}

// Create handles plain text POST requests to Create a new short URL.
func (s *Server) Create(c echo.Context) error {
	if c.Request().Header.Get("Content-Type") == "application/json" {
		return c.JSON(http.StatusBadRequest, "Bad request")
	}

	body := c.Request().Body
	url, err := io.ReadAll(body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	val := c.Get("userID")
	userID, ok := val.(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "userID is not of type int")
	}

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

// GetShort handles GET requests to redirect a short URL to its original target.
func (s *Server) GetShort(c echo.Context) error {
	short := c.Param("short")

	data, err := s.get(c, short)
	s.logger.Info("get data:", data, err)
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

// CreateJSON handles JSON POST requests to Create a new short URL.
func (s *Server) CreateJSON(c echo.Context) error {
	var (
		url requestJSON
		res responseJSON
	)

	err := c.Bind(&url)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}

	val := c.Get("userID")
	userID, ok := val.(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "userID is not of type int")
	}

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

// Ping is a health check endpoint that verifies database connectivity.
func (s *Server) Ping(c echo.Context) error {
	err := s.storage.HealthCheck()
	s.logger.Info(err)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Server error")
	}
	return c.JSON(http.StatusOK, nil)
}

// Batch handles Batch URL shortening requests and returns a list of results.
func (s *Server) Batch(c echo.Context) error {
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
		err  error
		data storage.Data
	)

	resp := make([]batchResponse, 0, len(reqs))

	val := c.Get("userID")
	userID, ok := val.(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "userID is not of type int")
	}

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
		s.logger.Infof("Batch response: %v", r)
	}

	s.logger.Infof("Batch response: %v", resp)
	return c.JSON(http.StatusCreated, resp)
}

// UserURLs returns all shortened URLs created by the current user.
func (s *Server) UserURLs(c echo.Context) error {
	val := c.Get("userID")
	userID, ok := val.(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "userID is not of type int")
	}

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

// DeleteUserURLs accepts a list of short URLs from the user and deletes them in background.
func (s *Server) DeleteUserURLs(c echo.Context) error {
	var shorts []string

	err := c.Bind(&shorts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "server error")
	}

	err = c.JSON(http.StatusAccepted, nil)
	if err != nil {
		s.logger.Error(err)
	}

	okShorts := make([]string, 0)
	val := c.Get("userID")
	userID, ok := val.(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, "userID is not of type int")
	}

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

	inCh := s.gen(okShorts...)
	ch1 := s.delete(inCh)
	ch2 := s.delete(inCh)
	for n := range s.fanIn(ch1, ch2) {
		s.logger.Info(n)
	}

	return nil
}

// Stats returns count urls and users
func (s *Server) Stats(c echo.Context) error {
	type response struct {
		URLS  int `json:"urls"`
		Users int `json:"users"`
	}

	var (
		resp response
		err  error
	)

	resp.URLS, resp.Users, err = s.storage.Stats()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "server error")
	}

	return c.JSON(http.StatusOK, &resp)
}
