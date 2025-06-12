package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// gzipResponseWriter wraps http.ResponseWriter and adds gzip compression support.
type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

// Write writes compressed data to the response using gzip.
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

// Header returns the header map that will be sent by WriteHeader.
func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// WriteHeader sends an HTTP response header with the provided status code.
func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

// gzipMiddleware is an Echo middleware that handles gzip compression and decompression.
// It compresses the response if the client supports gzip, and decompresses the request body
// if it's received in gzip format.
func (s *Server) gzipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		acceptEncoding := c.Request().Header.Get("Accept-Encoding")
		contentType := c.Response().Header().Get("Content-Type")
		if strings.Contains(acceptEncoding, "gzip") && (strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")) {
			c.Response().Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(c.Response().Writer)
			defer gz.Close()

			c.Response().Writer = &gzipResponseWriter{
				ResponseWriter: c.Response().Writer,
				writer:         gz,
			}
		}

		contentEncoding := c.Request().Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			reader, err := gzip.NewReader(c.Request().Body)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
			}
			defer reader.Close()
			c.Request().Body = io.NopCloser(reader)
		}

		return next(c)
	}
}
