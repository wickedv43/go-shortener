package url

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/wickedv43/go-shortener/internal/storage"
)

var (
	// ErrEmptyURL is returned when the URL is empty.
	ErrEmptyURL = errors.New("URL is empty")

	// ErrConflict is returned when a short URL already exists.
	ErrConflict = errors.New("conflict")

	// ErrBadRequest is returned when the request is invalid.
	ErrBadRequest = errors.New("bad request")

	// ErrInternal is returned when an internal error occurs.
	ErrInternal = errors.New("internal error")

	// ErrUnauthorized is returned when the user is not authorized.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrGone is returned when the URL has been deleted or is no longer available.
	ErrGone = errors.New("gone")

	// ErrNoContent is returned when no data is available for the user.
	ErrNoContent = errors.New("no content")
)

// Shortener defines the interface for the URL shortening service.
//
// The service allows saving, retrieving, deleting, and aggregating URLs by user.
// It also provides health checks and usage statistics.
type Shortener interface {
	// Save persists a new original URL for the specified user.
	// Returns the saved data or ErrConflict if the URL already exists.
	Save(ctx context.Context, originalURL string, userID int) (storage.Data, error)

	// Get retrieves URL data by its short identifier (shortID).
	Get(ctx context.Context, shortID string) (storage.Data, error)

	// GetAll retrieves all stored URLs for the specified user.
	// Returns ErrNoContent if no data is available.
	GetAll(ctx context.Context, userID int) ([]storage.Data, error)

	// BatchDelete deletes a batch of short URLs from storage.
	BatchDelete(shorts []string) error

	// DeleteUserURLS deletes the provided list of short URLs for the specified user.
	// Only URLs owned by the user and not already marked as deleted will be processed.
	DeleteUserURLS(ctx context.Context, userID int, shorts []string) error

	// Stats returns service usage statistics:
	// the total number of stored URLs and the number of registered users.
	Stats() (urls int, users int, err error)

	// Ping performs a health check of the underlying storage.
	// Returns an error if the storage is not healthy.
	Ping() error

	// Gen creates a channel that emits the provided short URLs for batch processing.
	Gen(shorts ...string) chan string

	// Delete consumes a channel of short URLs and performs batch deletion.
	// Returns a channel containing the results of the deletion.
	Delete(inCh chan string) chan string

	// FanIn merges multiple input channels into a single output channel.
	FanIn(chs ...chan string) chan string
}

// Save persists a new original URL or returns ErrConflict if it already exists.
func (u *URLService) Save(ctx context.Context, originalURL string, userID int) (storage.Data, error) {
	if originalURL == "" {
		return storage.Data{}, ErrEmptyURL
	}

	data, err := u.storage.Get(ctx, originalURL)
	if err != nil {
		data.OriginalURL = originalURL
		data.ShortURL = ShortURL()
		data.UUID = userID

		err = u.storage.Save(ctx, data)
		if err != nil {
			return storage.Data{}, errors.Wrap(err, "save error")
		}
		return data, nil
	}

	return data, ErrConflict
}

// Get retrieves URL data from storage by its short alias.
func (u *URLService) Get(ctx context.Context, shortID string) (storage.Data, error) {
	data, err := u.storage.Get(ctx, shortID)
	if err != nil {
		return storage.Data{}, errors.Wrap(err, "get error")
	}
	return data, nil
}

// GetAll retrieves all stored URLs belonging to the specified user.
//
// If no data is found, returns ErrNoContent.
func (u *URLService) GetAll(ctx context.Context, userID int) ([]storage.Data, error) {
	data, err := u.storage.GetAll(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "getAll error")
	}

	if len(data) == 0 {
		return []storage.Data{}, ErrNoContent
	}

	for i := range data {
		data[i].ShortURL = fmt.Sprintf("%s/%s", u.cfg.Server.FlagSuffixAddr, data[i].ShortURL)
	}

	return data, nil
}

// BatchDelete removes a batch of short URLs from storage.
func (u *URLService) BatchDelete(shorts []string) error {
	return u.storage.BatchDelete(shorts)
}

// DeleteUserURLS removes a list of short URLs for the given user.
//
// Only the user's own non-deleted URLs will be processed and removed.
func (u *URLService) DeleteUserURLS(ctx context.Context, userID int, shorts []string) error {
	var (
		okShorts = make([]string, 0)
		err      error
	)

	for _, short := range shorts {
		var d storage.Data
		d, err = u.Get(ctx, short)
		if err != nil {
			return errors.Wrap(err, "get error")
		}
		if !d.DeletedFlag && d.UUID == userID {
			okShorts = append(okShorts, short)
		}
	}

	inCh := u.Gen(okShorts...)
	ch1 := u.Delete(inCh)
	ch2 := u.Delete(inCh)
	for n := range u.FanIn(ch1, ch2) {
		u.logger.Info(n)
	}

	return nil
}

// Stats returns the current statistics of the service: number of URLs and users.
func (u *URLService) Stats() (urls int, users int, err error) {
	return u.storage.Stats()
}

// Ping performs a health check of the storage layer.
//
// Returns an error if the storage is not healthy.
func (u *URLService) Ping() error {
	return u.storage.HealthCheck()
}
