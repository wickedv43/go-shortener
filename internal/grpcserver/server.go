// Package grpcserver implements the gRPC server for the URL shortener service.
package grpcserver

import (
	"net"

	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/config"
	pb "github.com/wickedv43/go-shortener/internal/grpcapi"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/url"
	"google.golang.org/grpc"
)

// Server represents the gRPC server for the URL shortener service.
//
// It implements the pb.ShortenerServer interface and handles gRPC requests.
type Server struct {
	pb.UnimplementedShortenerServer

	// GRPC is the underlying gRPC server instance.
	GRPC *grpc.Server

	// URLService provides URL shortening functionality.
	URLService url.Shortener

	cfg *config.Config
	log *logrus.Entry
}

// NewServer creates and returns a new gRPC Server instance.
//
// It initializes interceptors, configures the server, and registers gRPC services.
func NewServer(i do.Injector) (*Server, error) {
	s := &Server{
		URLService: do.MustInvoke[url.Shortener](i),
		cfg:        do.MustInvoke[*config.Config](i),
		log:        do.MustInvoke[*logger.Logger](i).WithField("component", "grpc"),
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			ChainUnaryInterceptors(
				s.LogUnaryInterceptor(),
				s.AuthInterceptor(),
				s.TrustedSubnetInterceptor(),
			),
		),
	)

	s.GRPC = grpcServer

	pb.RegisterShortenerServer(s.GRPC, s)

	return s, nil
}

// Start runs the gRPC server and begins listening for incoming connections.
//
// The server listens on TCP port 8081.
func (s *Server) Start() {
	s.log.Infof("Starting gRPC server")
	listen, err := net.Listen("tcp", ":8081")
	if err != nil {
		s.log.WithError(err).Fatal("failed to listen")
	}

	if err = s.GRPC.Serve(listen); err != nil {
		s.log.WithError(err).Fatal("failed to serve")
	}
}

// Shutdown gracefully stops the gRPC server.
//
// This method ensures that all ongoing requests are properly completed before shutdown.
func (s *Server) Shutdown() {
	s.log.Infof("gRPC shutdown complete ")
	s.GRPC.GracefulStop()
}
