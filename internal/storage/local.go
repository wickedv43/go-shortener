package storage

import (
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

func (l *LocalStorage) Save(d Data) error {
	l.locMem = append(l.locMem, d)

	return nil
}

func (l *LocalStorage) Get(s string) (Data, error) {
	for _, data := range l.locMem {
		if data.ShortURL == s || data.OriginalURL == s {
			return data, nil
		}
	}

	return Data{}, errors.New("not found")
}

func (l *LocalStorage) Delete(url string) error {
	for i, data := range l.locMem {
		if data.OriginalURL == url {
			l.locMem = append(l.locMem[:i], l.locMem[i+1:]...)
		}

	}
	return nil
}

func (l *LocalStorage) Load() error {
	//TODO implement me
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
