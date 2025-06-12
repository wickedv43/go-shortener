package server

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/wickedv43/go-shortener/internal/storage"
	"github.com/wickedv43/go-shortener/internal/url"
)

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
		return c.JSON(http.StatusBadRequest, url.ErrBadRequest.Error())
	}

	body := c.Request().Body
	original, err := io.ReadAll(body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
	}

	val := c.Get("userID")
	userID, ok := val.(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, url.ErrUnauthorized.Error())
	}

	data, err := s.URLService.Save(c.Request().Context(), string(original), userID)
	resURL := fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, data.ShortURL)
	if err != nil {
		if errors.Is(err, url.ErrConflict) {
			c.Response().WriteHeader(http.StatusConflict)
			_, err = c.Response().Write([]byte(resURL))
			return err
		}
		return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
	}

	c.Response().Header().Set("Content-Type", "text/plain")
	c.Response().WriteHeader(http.StatusCreated)
	_, err = c.Response().Write([]byte(resURL))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
	}

	return nil
}

// GetShort handles GET requests to redirect a short URL to its original target.
func (s *Server) GetShort(c echo.Context) error {
	short := c.Param("short")

	data, err := s.URLService.Get(c.Request().Context(), short)
	s.logger.Info("get data:", data, err)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
	}

	if data.DeletedFlag {
		return c.JSON(http.StatusGone, url.ErrGone.Error())
	}

	c.Response().Header().Set("Location", data.OriginalURL)
	c.Response().WriteHeader(http.StatusTemporaryRedirect)
	return nil
}

// CreateJSON handles JSON POST requests to Create a new short URL.
func (s *Server) CreateJSON(c echo.Context) error {
	var (
		req requestJSON
		res responseJSON
	)

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
	}

	val := c.Get("userID")
	userID, ok := val.(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, url.ErrUnauthorized.Error())
	}

	data, err := s.URLService.Save(c.Request().Context(), req.URL, userID)
	res.Result = fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, data.ShortURL)
	if err != nil {
		if errors.Is(err, url.ErrConflict) {
			return c.JSON(http.StatusConflict, res.Result)
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
	}

	return c.JSON(http.StatusCreated, res)
}

// Ping is a health check endpoint that verifies database connectivity.
func (s *Server) Ping(c echo.Context) error {
	err := s.URLService.Ping()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
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
		return c.JSON(http.StatusUnauthorized, url.ErrUnauthorized.Error())
	}

	err = c.Bind(&reqs)
	if err != nil {
		return c.JSON(http.StatusBadRequest, url.ErrBadRequest.Error())
	}

	if len(reqs) == 0 {
		return c.JSON(http.StatusBadRequest, url.ErrBadRequest.Error())
	}

	for _, req := range reqs {
		data, err = s.URLService.Save(c.Request().Context(), req.OriginalURL, userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
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
		return c.JSON(http.StatusUnauthorized, url.ErrUnauthorized.Error())
	}

	urls, err := s.URLService.GetAll(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, url.ErrNoContent) {
			return c.JSON(http.StatusNoContent, url.ErrNoContent.Error())
		}
		return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
	}

	return c.JSON(http.StatusOK, urls)
}

// DeleteUserURLs accepts a list of short URLs from the user and deletes them in background.
func (s *Server) DeleteUserURLs(c echo.Context) error {
	var shorts []string

	err := c.Bind(&shorts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
	}

	if len(shorts) == 0 {
		err = c.JSON(http.StatusAccepted, nil)
		if err != nil {
			s.logger.Error(err)
		}
		return nil
	}

	err = c.JSON(http.StatusAccepted, nil)
	if err != nil {
		s.logger.Error(err)
	}

	val := c.Get("userID")
	userID, ok := val.(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, url.ErrUnauthorized.Error())
	}

	err = s.URLService.DeleteUserURLS(c.Request().Context(), userID, shorts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
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

	resp.URLS, resp.Users, err = s.URLService.Stats()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, url.ErrInternal.Error())
	}

	return c.JSON(http.StatusOK, &resp)
}
