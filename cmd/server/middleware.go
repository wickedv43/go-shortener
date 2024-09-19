package server

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

func (s *Server) logHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		// before request

		c.Next()

		// after request
		latency := time.Since(t)

		reqMethod := c.Request.Method
		reqURI := c.Request.RequestURI

		respStatus := c.Writer.Status()
		respSize := c.Writer.Size()

		s.logger.WithFields(logrus.Fields{
			"method":  reqMethod,
			"uri":     reqURI,
			"latency": latency,
		}).Infoln("request")

		s.logger.WithFields(logrus.Fields{
			"size":   respSize,
			"status": respStatus,
		}).Info("response")
	}
}

type gzipWriter struct {
	gin.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (s *Server) gzipHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		if strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") ||
			strings.Contains(c.Request.Header.Get("Content-Type"), "text/html") {
			gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			defer gz.Close()

			c.Header("Content-Encoding", "gzip")

			c.Writer = gzipWriter{c.Writer, gz}
		}

		c.Next()
	}
}
