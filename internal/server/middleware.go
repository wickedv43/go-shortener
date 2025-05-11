package server

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// logHandler is a middleware that logs details about each HTTP request,
// including method, URI, response status, size, and processing latency.
func (s *Server) logHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		t := time.Now()

		err := next(c)
		if err != nil {
			return err
		}

		latency := time.Since(t)
		reqMethod := c.Request().Method
		reqURI := c.Request().RequestURI
		respStatus := c.Response().Status
		respSize := c.Response().Size

		s.logger.WithFields(logrus.Fields{
			"method":      reqMethod,
			"uri":         reqURI,
			"latency":     latency,
			"resp_size":   respSize,
			"resp_status": respStatus,
		}).Infoln("request")

		return nil
	}
}

// CORSMiddleware is a middleware that adds CORS headers to the response,
// allowing cross-origin requests from any domain and handling preflight OPTIONS requests.
func (s *Server) CORSMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", "*")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Response().Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request().Method == "OPTIONS" {
			return c.JSON(http.StatusNoContent, "")
		}

		return next(c)
	}
}
