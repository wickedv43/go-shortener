package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func (w *gzipResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (s *Server) gzipMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		acceptEncoding := c.Request().Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
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
