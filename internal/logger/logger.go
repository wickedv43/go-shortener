package logger

import (
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

func NewLogger(i do.Injector) (*Logger, error) {
	logger, err := do.InvokeStruct[Logger](i)
	if err != nil {
		return nil, errors.Wrapf(err, "invoke logger")
	}

	logger.Logger = logrus.New()
	logger.Logger.SetLevel(logrus.InfoLevel)

	return logger, nil
}
