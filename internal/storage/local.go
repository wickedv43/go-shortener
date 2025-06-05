package storage

import (
	"context"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/logger"
)

// LocalStorage is an in-memory implementation of the DataKeeper interface.
// It is mainly used for development or testing without a persistent backend.
type LocalStorage struct {
	log    *logrus.Entry // Logger instance scoped to local storage.
	LocMem []Data        // In-memory slice storing Data objects.
}

// NewLocalStorage creates and initializes a LocalStorage instance using dependency injection.
func NewLocalStorage(i do.Injector) (*LocalStorage, error) {
	storage, err := do.InvokeStruct[LocalStorage](i)
	log := do.MustInvoke[*logger.Logger](i).WithField("component", "db")

	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	locMem := make([]Data, 0)

	storage.LocMem = locMem
	storage.log = log

	return storage, nil
}

// Save stores the given Data object into the in-memory slice.
func (l *LocalStorage) Save(_ context.Context, d Data) error {
	l.LocMem = append(l.LocMem, d)
	return nil
}

// Get retrieves a Data object by matching its short or original URL.
func (l *LocalStorage) Get(_ context.Context, s string) (Data, error) {
	for _, data := range l.LocMem {
		if data.ShortURL == s || data.OriginalURL == s {
			return data, nil
		}
	}
	return Data{}, errors.New("not found")
}

// GetAll returns all URL records associated with the given user ID.
func (l *LocalStorage) GetAll(_ context.Context, userID int) ([]Data, error) {
	data := make([]Data, 0)
	for _, d := range l.LocMem {
		if d.UUID == userID {
			data = append(data, d)
		}
	}

	if len(data) == 0 {
		return data, errors.New("no content")
	}

	return data, nil
}

// BatchDelete is a stub implementation that does nothing for in-memory storage.
func (l *LocalStorage) BatchDelete(_ []string) error {
	return nil
}

// HealthCheck is a stub implementation and always returns nil.
func (l *LocalStorage) HealthCheck() error {
	// Not applicable for in-memory storage.
	return nil
}

// Close is a stub method for interface compatibility; no resources to clean up.
func (l *LocalStorage) Close() error {
	// Not applicable for in-memory storage.
	return nil
}

func (l *LocalStorage) Shutdown(_ context.Context) error {
	l.log.Info("Shutting down local storage (no-op)")
	return nil
}
