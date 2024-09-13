package storage

import (
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/cmd/logger"
)

type Storage struct {
	db     map[string]string
	logger *logrus.Entry
}

func NewStorage(i do.Injector) (*Storage, error) {
	storage, err := do.InvokeStruct[Storage](i)
	log := do.MustInvoke[*logger.Logger](i).WithField("component", "db")

	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	db := make(map[string]string)
	storage.db = db
	storage.logger = log

	return storage, err
}
