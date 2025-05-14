// Package storage defines interfaces and data structures for working with persistent storage.
// It provides an abstraction layer to store, retrieve, and manage shortened URL data.
package storage

import (
	"context"

	_ "github.com/lib/pq" // PostgreSQL driver import for side effects.
)

// Data represents a record of a shortened URL, including the user ID,
// original URL, short alias, and deletion flag.
type Data struct {
	UUID        int    `json:"uuid"`         // ID of the user who created the short URL.
	ShortURL    string `json:"short_url"`    // The generated short URL alias.
	OriginalURL string `json:"original_url"` // The original long URL.
	DeletedFlag bool   `json:"is_deleted"`   // Flag indicating whether the URL has been deleted.
}

// DataKeeper defines the interface for interacting with the storage layer.
// It includes methods for saving, retrieving, listing, deleting, and health-checking storage.
type DataKeeper interface {
	// Save stores a new shortened URL entry.
	Save(c context.Context, d Data) error

	// Get retrieves a single URL entry by either short or original URL.
	Get(c context.Context, s string) (Data, error)

	// GetAll returns all stored entries associated with a specific user.
	GetAll(c context.Context, userID int) ([]Data, error)

	// BatchDelete removes multiple short URLs in a single operation.
	BatchDelete(short []string) error

	// HealthCheck verifies connectivity with the storage backend.
	HealthCheck() error

	// Close releases any resources associated with the storage (e.g. DB connection).
	Close() error
}
