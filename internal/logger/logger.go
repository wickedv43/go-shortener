package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

var log = logrus.New()

func Logger() gin.HandlerFunc {
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

		log.WithFields(logrus.Fields{
			"method":  reqMethod,
			"uri":     reqURI,
			"latency": latency,
		}).Infoln("req")

		log.WithFields(logrus.Fields{
			"size":   respSize,
			"status": respStatus,
		}).Info("resp")

	}
}
