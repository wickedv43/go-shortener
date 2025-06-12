package grpc

import (
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"google.golang.org/grpc"

	"github.com/wickedv43/go-shortener/internal/url"
)

type Server struct {
	UnimplementedShortenerServer

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
	return s, nil
}

func (s *Server) Start() {

}
