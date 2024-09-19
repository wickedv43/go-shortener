package server

import (
	"bufio"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"io"
	"net"
	"net/http"
)

// w - gin
// zw - gzip
type gzipWriter struct {
	w  gin.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(c *gin.Context) *gzipWriter {
	return &gzipWriter{
		w:  c.Writer,
		zw: gzip.NewWriter(c.Writer),
	}
}

func (g *gzipWriter) Header() http.Header {
	return g.w.Header()
}

func (g *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		g.w.Header().Set("Content-Encoding", "gzip")
	}
	g.w.WriteHeader(statusCode)
}

func (g *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return g.w.Hijack()
}

func (g *gzipWriter) Flush() {
	g.w.Flush()
}

func (g *gzipWriter) CloseNotify() <-chan bool {
	return g.w.CloseNotify()
}

func (g *gzipWriter) Status() int {
	return g.w.Status()
}

func (g *gzipWriter) Size() int {
	return g.w.Size()
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.w.Write([]byte(s))
}

func (g *gzipWriter) Written() bool {
	return g.w.Written()
}

func (g *gzipWriter) WriteHeaderNow() {
	g.w.WriteHeaderNow()
}

func (g *gzipWriter) Pusher() http.Pusher {
	return g.w.Pusher()
}

func (g *gzipWriter) Write(p []byte) (int, error) {
	return g.zw.Write(p)
}

func (g *gzipWriter) Close() error {
	return g.zw.Close()
}

type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

func (g *gzipReader) Read(p []byte) (n int, err error) {
	return g.zr.Read(p)
}

func (g *gzipReader) Close() error {
	if err := g.r.Close(); err != nil {
		return err
	}
	return g.zr.Close()
}
