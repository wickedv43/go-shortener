package server

import (
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
)

type Server struct {
	engine  *gin.Engine
	cfg     *config.Config
	storage *storage.Storage
	logger  *logrus.Entry
}

func NewServer(i do.Injector) (*Server, error) {
	cfg := do.MustInvoke[*config.Config](i)
	stor := do.MustInvoke[*storage.Storage](i)
	lg := do.MustInvoke[*logger.Logger](i).WithField("component", "gin")

	server, err := do.InvokeStruct[Server](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct error")
	}

	e := gin.New()
	e.Use(gin.Recovery(), server.logHandler(), server.gzipMiddleware(), server.CORSMiddleware())

	server.engine = e
	server.cfg = cfg
	server.storage = stor
	server.logger = lg

	server.engine.POST(`/`, server.addNew)
	server.engine.POST(`/api/shorten`, server.addNewJSON)
	server.engine.GET(`/:short`, server.getShort)

	return server, nil
}

func (s *Server) Start() {
	err := s.storage.LoadFromFile()
	if err != nil {
		s.logger.Fatal(errors.Wrapf(err, "load file: %s", s.cfg.Server.FlagStoragePath))
	}

	s.logger.Info("server started")
	err = s.engine.Run(s.cfg.Server.FlagRunAddr)
	if err != nil {
		s.logger.Fatal(errors.Wrap(err, "start server"))
	}
}
