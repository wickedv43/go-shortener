// Package server contains the main HTTP server logic for the URL shortener service.
// It defines middleware, routing, authentication, compression, logging, CORS handling,
// and endpoint handlers for operations such as creating, retrieving, and deleting short URLs.
// The server uses Echo as its web framework and integrates with pprof for profiling,
// as well as JWT-based user identification via cookies.
package server

import (
	"context"
	"net/http"
	"net/http/pprof"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/url"
)

// Server represents the HTTP server and its dependencies.
type Server struct {
	echo       *echo.Echo     // Echo HTTP server instance.
	cfg        *config.Config // Configuration settings.
	URLService url.Shortener
	logger     *logrus.Entry // Structured logger.
}

// NewServer creates and configures a new Server instance using dependency injection.
// It sets up middlewares, profiling endpoints, and all routes.
func NewServer(i do.Injector) (*Server, error) {
	s, err := do.InvokeStruct[Server](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct error")
	}

	s.echo = echo.New()
	s.echo.Use(middleware.Recover(), s.gzipMiddleware, s.authMiddleware, s.logHandler, s.CORSMiddleware)

	s.URLService = do.MustInvoke[url.Shortener](i)

	s.cfg = do.MustInvoke[*config.Config](i)
	s.logger = do.MustInvoke[*logger.Logger](i).WithField("component", "server")

	// pprof endpoints
	s.echo.GET("/debug/pprof/", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
	s.echo.GET("/debug/pprof/cmdline", echo.WrapHandler(http.HandlerFunc(pprof.Cmdline)))
	s.echo.GET("/debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	s.echo.GET("/debug/pprof/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	s.echo.POST("/debug/pprof/symbol", echo.WrapHandler(http.HandlerFunc(pprof.Symbol)))
	s.echo.GET("/debug/pprof/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	s.echo.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))

	// API routes
	s.echo.POST(`/`, s.Create)
	s.echo.GET(`/:short`, s.GetShort)
	s.echo.GET(`/ping`, s.Ping)
	s.echo.POST(`/api/shorten`, s.CreateJSON)
	s.echo.POST(`/api/shorten/batch`, s.Batch)
	s.echo.GET(`/api/user/urls`, s.UserURLs)
	s.echo.DELETE(`/api/user/urls`, s.DeleteUserURLs)

	// trusted subnet?
	trustedIP := s.echo.Group("")
	trustedIP.Use(s.TrustedSubnetMiddleware)
	trustedIP.GET(`/api/internal/stats`, s.Stats)

	return s, nil
}

// Start runs the HTTP or HTTPS server on the configured address.
func (s *Server) Start() {
	s.logger.Info("starting server...")

	if s.cfg.Server.FlagHTTPS {
		err := s.echo.StartTLS(s.cfg.Server.FlagRunAddr, s.cfg.Server.FlagCertPath, s.cfg.Server.FlagKeyPath)
		if errors.Is(err, http.ErrServerClosed) {
			s.logger.Info("echo shutdown complete")
		} else {
			s.logger.Error(err, "start http server")
		}
	} else {
		err := s.echo.Start(s.cfg.Server.FlagRunAddr)
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				s.logger.Info("echo shutdown complete")
			} else {
				s.logger.Error(err, "start http server")
			}

		}
	}
}

// Shutdown server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}
