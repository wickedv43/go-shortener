package storage

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Put(d Data) - saves Data in local memory and file
func (s *Storage) Put(d Data) error {
	d.UUID = uuid.New().ClockSequence()

	err := s.Save(d)
	if err != nil {
		return errors.Wrap(err, "save")
	}

	return err
}

// Get(short string) - get data from local memory
func (s *Storage) getFromLocMem(short string) (string, bool) {
	var url string

	for _, d := range s.db {
		if d.ShortURL == short {
			url = d.OriginalURL

			return url, true
		}
	}
	return url, false
}

// InStorage(url string) - check if extended url is already in the database
func (s *Storage) InStorage(url string) (string, bool) {
	if s.isPostgresAvailable() {
		short, found, err := s.checkPostgres(url)
		if err != nil {
			s.log.WithError(err).Error("Error while checking URL in PostgreSQL")
		}
		if found {
			return short, true
		}
	}

	//locMem
	for _, d := range s.db {
		if d.OriginalURL == url {
			return d.ShortURL, true
		}
	}

	return "", false
}

func (s *Storage) Get(short string) (string, bool) {
	if s.isPostgresAvailable() {
		url, found := s.getFromPostgres(short)
		if found {
			return url, true
		}
	}
	return s.getFromLocMem(short)
}
