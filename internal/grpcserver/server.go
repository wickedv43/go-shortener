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

type Server struct {
	pb.UnimplementedShortenerServer

	GRPC       *grpc.Server
	URLService url.Shortener

	cfg *config.Config
	log *logrus.Entry
}

func NewServer(i do.Injector) (*Server, error) {
	s := &Server{
		GRPC:       grpc.NewServer(),
		URLService: do.MustInvoke[url.Shortener](i),
		cfg:        do.MustInvoke[*config.Config](i),
		log:        do.MustInvoke[*logger.Logger](i).WithField("component", "grpc"),
	}

	pb.RegisterShortenerServer(s.GRPC, s)

	return s, nil
}

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

func (s *Server) Shutdown() {
	s.log.Infof("gRPC shutdown complete ")
	s.GRPC.GracefulStop()
}
