package server

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/wickedv43/go-shortener/cmd/config"
	"github.com/wickedv43/go-shortener/cmd/storage"
	"log"
)

type Server struct {
	engine  *gin.Engine
	cfg     *config.Config
	storage *storage.Storage
}

func NewServer(i do.Injector) (*Server, error) {
	cfg := do.MustInvoke[*config.Config](i)
	stor := do.MustInvoke[*storage.Storage](i)

	server, err := do.InvokeStruct[Server](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct error")
	}

	e := gin.New()
	e.Use(gin.Recovery(), gin.Logger())

	server.engine = e
	server.cfg = cfg
	server.storage = stor

	server.engine.POST(`/`, server.addNew)
	server.engine.GET(`/:short`, server.getShort)

	return server, nil
}

func (s *Server) Start() {
	err := s.engine.Run(s.cfg.FlagRunAddr)
	if err != nil {
		log.Fatal(errors.Wrap(err, "start server"))
	}
}
