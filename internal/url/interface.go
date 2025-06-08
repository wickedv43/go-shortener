package url

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/wickedv43/go-shortener/internal/storage"
)

var (
	ErrEmptyURL = errors.New("URL is empty")
	// ErrConflict is returned when a short URL already exists.
	ErrConflict     = errors.New("conflict")
	ErrBadRequest   = errors.New("bad request")
	ErrInternal     = errors.New("internal error")
	ErrUnauthorized = errors.New("unauthorized")
	ErrGone         = errors.New("gone")
	// ErrNoContent is returned when no data is available for the user.
	ErrNoContent = errors.New("no content")
)

// TODO: godoc comma
type Shortener interface {
	Save(ctx context.Context, originalURL string, userID int) (storage.Data, error)
	Get(ctx context.Context, shortID string) (storage.Data, error)
	GetAll(ctx context.Context, userID int) ([]storage.Data, error)
	DeleteBatch(shorts []string) error
	DeleteUserURLS(ctx context.Context, userID int, shorts []string) error

	HealthCheck() error
	Stats() (urls int, users int, err error)

	Gen(shorts ...string) chan string
	Delete(inCh chan string) chan string
	FanIn(chs ...chan string) chan string
}

// Save persists a new URL or returns ErrConflict if it already exists.
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
func (u *URLService) GetAll(ctx context.Context, userID int) ([]storage.Data, error) {
	data, err := u.storage.GetAll(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "getAll error")
	}

	if len(data) == 0 {
		return []storage.Data{}, ErrNoContent
	}

	for _, d := range data {
		d.ShortURL = fmt.Sprintf("%s/%s", u.cfg.Server.FlagSuffixAddr, d.ShortURL)
	}

	return data, nil
}

// BatchDelete removes a Batch of short URLs from storage.
func (u *URLService) DeleteBatch(shorts []string) error {
	return u.storage.BatchDelete(shorts)
}

func (u *URLService) DeleteUserURLS(ctx context.Context, userID int, shorts []string) error {

	var (
		okShorts = make([]string, 0)

		err error
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

func (u *URLService) Stats() (urls int, users int, err error) {
	return u.storage.Stats()
}
func (u *URLService) HealthCheck() error {
	return u.storage.HealthCheck()
}
