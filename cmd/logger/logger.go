package logger

import (
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/cmd/config"
)

type Logger struct {
	*logrus.Logger
}

func NewLogger(i do.Injector) (*Logger, error) {
	cfg := do.MustInvoke[*config.Config](i)

	log := logrus.New()
	log.SetLevel(cfg.Logger.Lvl)

	return &Logger{
		log,
	}, nil
}
