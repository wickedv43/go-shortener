package storage

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// uuidCounter() - for uuid count Data
func (s *Storage) uuidCounter() int {
	counter := 1
	for _, d := range s.db {
		if d.UUID >= counter {
			counter = d.UUID + 1
		}
	}
	return counter
}

// Put(d Data) - saves Data in local memory and file
func (s *Storage) Put(d Data) {
	d.UUID = s.uuidCounter()
	s.db = append(s.db, d)

	err := s.Save(d)
	if err != nil {
		s.log.Fatal(errors.Wrap(err, "save storage"))
	}

	s.log.WithFields(logrus.Fields{
		"url":   d.OriginalURL,
		"short": d.ShortURL,
	}).Infoln("saved")
}

// Get(short string) - get data from local memory
func (s *Storage) Get(short string) (string, bool) {
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
