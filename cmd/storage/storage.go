package storage

import (
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
)

type Storage struct {
	db map[string]string
}

func NewStorage(i do.Injector) (*Storage, error) {
	storage, err := do.InvokeStruct[Storage](i)

	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	db := make(map[string]string)
	storage.db = db

	return storage, err
}
