package storage

import (
	_ "github.com/lib/pq"
)

type Data struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type DataKeeper interface {
	Save(d Data) error
	Get(s string) (Data, error)
	Delete(string) error

	HealthCheck() error

	Close() error
}
