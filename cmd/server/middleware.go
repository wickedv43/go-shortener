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

func (s Server) gzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		ow := c.Writer

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newGZIPWriter(c)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
}
