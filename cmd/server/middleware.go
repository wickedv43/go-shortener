package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

func (s *Server) gzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			gz := newCompressWriter(c)

			defer gz.Close()

			c.Writer = gz
		}

		contentEncoding := c.Request.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			decompresedBody, err := newCompressReader(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}

			c.Request.Body = decompresedBody
			defer decompresedBody.Close()
		}

		c.Next()
	}
}
