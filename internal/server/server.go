package server

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
	"github.com/wickedv43/go-shortener/internal/storage"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
)

type Server struct {
	echo    *echo.Echo
	cfg     *config.Config
	storage storage.DataKeeper
	logger  *logrus.Entry
}

func NewServer(i do.Injector) (*Server, error) {
	s, err := do.InvokeStruct[Server](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct error")
	}

	s.echo = echo.New()
	s.echo.Use(middleware.Recover(), middleware.Gzip(), s.logHandler, s.CORSMiddleware)

	s.cfg = do.MustInvoke[*config.Config](i)
	s.logger = do.MustInvoke[*logger.Logger](i).WithField("component", "server")

	s.storage = s.SelectStorage(i)

	s.echo.POST(`/`, s.create)
	s.echo.POST(`/api/shorten`, s.createJSON)
	s.echo.POST(`api/shorten/batch`, s.batch)
	s.echo.GET(`/:short`, s.getShort)
	s.echo.GET(`/ping`, s.ping)

	return s, nil
}

func (s *Server) SelectStorage(i do.Injector) storage.DataKeeper {
	var err error

	if _, err = do.Invoke[*storage.PostgresStorage](i); err == nil {
		s.logger.WithField("component", "postgres").Debug("using postgres storage")
		return do.MustInvoke[*storage.PostgresStorage](i)
	}

	//Пробуем файловое хранилище
	if _, err = do.Invoke[*storage.FileStorage](i); err == nil {
		s.logger.WithField("storage", i).Info("using file storage")
		return do.MustInvoke[*storage.FileStorage](i)
	}

	s.logger.WithField("storage", "locMem").Info("using memory storage")
	return do.MustInvoke[*storage.LocalStorage](i)
}

func (s *Server) save(expand string) (storage.Data, error) {
	if expand == "" {
		return storage.Data{}, errors.New("empty url")
	}

	data, err := s.storage.Get(expand)

	if err != nil {
		data.OriginalURL = expand
		data.ShortURL = ShortURL()
		data.UUID = uuid.New().ClockSequence()

		err = s.storage.Save(data)
		if err != nil {
			return storage.Data{}, errors.Wrap(err, "save error")
		}
		return data, nil
	}

	return data, errConflict
}

func (s *Server) get(short string) (storage.Data, error) {
	data, err := s.storage.Get(short)
	if err != nil {
		return storage.Data{}, errors.Wrap(err, "get error")
	}
	return data, nil
}

func (s *Server) Start() {
	s.logger.Info("server started")
	err := s.echo.Start(s.cfg.Server.FlagRunAddr)
	if err != nil {
		s.logger.Fatal(errors.Wrap(err, "start server"))
	}
}
