package storage

import (
	"context"

	_ "github.com/lib/pq"
)

type Data struct {
	UUID        int    `json:"-"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type DataKeeper interface {
	Save(c context.Context, d Data) error
	Get(c context.Context, s string) (Data, error)
	GetAll(c context.Context, userID int) ([]Data, error)
	Delete(c context.Context, s string) error

	HealthCheck() error

	Close() error
}
