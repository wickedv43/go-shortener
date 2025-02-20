package server

import (
	"context"
	"fmt"

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

	//channels
	urlDeleteChan chan string
}

func NewServer(i do.Injector) (*Server, error) {
	s, err := do.InvokeStruct[Server](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct error")
	}

	s.echo = echo.New()
	s.echo.Use(middleware.Recover(), s.gzipMiddleware, s.authMiddleware, s.logHandler, s.CORSMiddleware)

	s.cfg = do.MustInvoke[*config.Config](i)
	s.logger = do.MustInvoke[*logger.Logger](i).WithField("component", "server")

	s.storage = s.selectStorage(i)

	//routes
	s.echo.POST(`/`, s.create)
	s.echo.GET(`/:short`, s.getShort)

	s.echo.GET(`/ping`, s.ping)

	s.echo.POST(`/api/shorten`, s.createJSON)
	s.echo.POST(`/api/shorten/batch`, s.batch)

	s.echo.GET(`/api/user/urls`, s.userURLs)
	s.echo.DELETE(`/api/user/urls`, s.deleteUserURLs)

	//channels
	s.urlDeleteChan = make(chan string, 100)

	//workers
	go s.deleteWorker()

	return s, nil
}

func (s *Server) selectStorage(i do.Injector) storage.DataKeeper {
	var err error

	if _, err = do.Invoke[*storage.PostgresStorage](i); err == nil {
		s.logger.WithField("component", "postgres").Debug("using postgres storage")
		return do.MustInvoke[*storage.PostgresStorage](i)
	}

	//Пробуем файловое хранилище
	if _, err = do.Invoke[*storage.FileStorage](i); err == nil {
		s.logger.WithField("storage", "file").Info("using file storage")
		return do.MustInvoke[*storage.FileStorage](i)
	}

	s.logger.WithField("storage", "locMem").Info("using memory storage")
	return do.MustInvoke[*storage.LocalStorage](i)
}

func (s *Server) save(ctx context.Context, expand string, userID int) (storage.Data, error) {
	if expand == "" {
		return storage.Data{}, errors.New("empty url")
	}

	data, err := s.storage.Get(ctx, expand)

	if err != nil {
		data.OriginalURL = expand
		data.ShortURL = ShortURL()
		data.UUID = userID

		err = s.storage.Save(ctx, data)
		if err != nil {
			return storage.Data{}, errors.Wrap(err, "save error")
		}
		return data, nil
	}

	return data, ErrConflict
}

func (s *Server) get(c echo.Context, short string) (storage.Data, error) {
	ctx := c.Request().Context()

	data, err := s.storage.Get(ctx, short)
	if err != nil {
		return storage.Data{}, errors.Wrap(err, "get error")
	}
	return data, nil
}

func (s *Server) getAll(c echo.Context, userID int) ([]storage.Data, error) {
	ctx := c.Request().Context()

	data, err := s.storage.GetAll(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "getAll error")
	}

	if len(data) == 0 {
		return []storage.Data{}, ErrNoContent
	}

	for _, d := range data {
		d.ShortURL = fmt.Sprintf("%s/%s", s.cfg.Server.FlagSuffixAddr, d.ShortURL)
	}

	return data, nil
}

func (s *Server) batchDelete(shorts []string) error {
	return s.storage.BatchDelete(shorts)
}

func (s *Server) Start() {
	s.logger.Info("server started")
	err := s.echo.Start(s.cfg.Server.FlagRunAddr)
	if err != nil {
		s.logger.Fatal(errors.Wrap(err, "start server"))
	}
}
