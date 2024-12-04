package storage

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Put(d Data) - saves Data in local memory and file
func (s *Storage) Put(d Data) {
	d.UUID = uuid.New().ClockSequence()

	err := s.Save(d)
	if err != nil {
		s.log.Fatal(errors.Wrap(err, "save"))
	}
}

// Get(short string) - get data from local memory
func (s *Storage) getFromFile(short string) (string, bool) {
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
	var short string

	for _, d := range s.db {
		if d.OriginalURL == url {
			short = d.ShortURL
			return short, true
		}
	}
	return short, false
}

func (s *Storage) Get(short string) (string, bool) {
	if s.isPostgresAvailable() {
		url, found := s.getFromPostgres(short)
		if found {
			return url, true
		}
	}
	return s.getFromFile(short)
}
