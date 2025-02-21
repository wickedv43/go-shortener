package storage

import (
	"context"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/logger"
)

type LocalStorage struct {
	locMem []Data
	log    *logrus.Entry
}

func NewLocalStorage(i do.Injector) (*LocalStorage, error) {
	storage, err := do.InvokeStruct[LocalStorage](i)
	log := do.MustInvoke[*logger.Logger](i).WithField("component", "db")

	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	// locMem database
	locMem := make([]Data, 0)

	storage.locMem = locMem
	storage.log = log

	return storage, nil
}

func (l *LocalStorage) Save(_ context.Context, d Data) error {
	l.locMem = append(l.locMem, d)

	return nil
}

func (l *LocalStorage) Get(_ context.Context, s string) (Data, error) {
	for _, data := range l.locMem {
		if data.ShortURL == s || data.OriginalURL == s {
			return data, nil
		}
	}

	return Data{}, errors.New("not found")
}

func (l *LocalStorage) GetAll(_ context.Context, userID int) ([]Data, error) {
	data := make([]Data, 0)

	for _, d := range l.locMem {
		if d.UUID == userID {
			data = append(data, d)
		}
	}

	if len(data) == 0 {
		return data, errors.New("no content")
	}

	return data, nil
}

// bad var for in for
func (l *LocalStorage) BatchDelete(short ...string) error {
	//for _, short := range shorts {
	//	for i, data := range l.locMem {
	//		if !data.DeletedFlag {
	//			if data.OriginalURL == short || data.ShortURL == short {
	//				l.locMem[i].DeletedFlag = true
	//			}
	//		}
	//
	//	}
	//}

	return nil
}

func (l *LocalStorage) HealthCheck() error {
	//TODO implement me
	return nil
}

func (l *LocalStorage) Close() error {
	//TODO implement me
	return nil
}
