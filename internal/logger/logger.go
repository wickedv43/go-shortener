// Package logger provides a wrapper around logrus for centralized logging.
package logger

import (
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
)

// Logger is a wrapper around logrus.Logger used throughout the project.
type Logger struct {
	*logrus.Logger // Embedded logrus logger instance.
}

// NewLogger creates and initializes a Logger with InfoLevel logging.
// It uses the samber/do dependency injection container to resolve dependencies.
func NewLogger(i do.Injector) (*Logger, error) {
	logger, err := do.InvokeStruct[Logger](i)
	if err != nil {
		return nil, errors.Wrapf(err, "invoke logger")
	}

	logger.Logger = logrus.New()
	logger.Logger.SetLevel(logrus.InfoLevel)

	return logger, nil
}
