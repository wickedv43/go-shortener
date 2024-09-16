package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

func (s *Server) HandlerLog() gin.HandlerFunc {
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
