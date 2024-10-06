package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

func (s *Server) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
