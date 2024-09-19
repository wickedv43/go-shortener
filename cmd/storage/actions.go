package storage

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *Storage) uuidCounter() int {
	counter := 1
	for _, d := range s.db {
		if d.UUID >= counter {
			counter = d.UUID + 1
		}
	}
	return counter
}

func (s *Storage) Put(d Data) {
	d.UUID = s.uuidCounter()
	s.db = append(s.db, d)

	err := s.Save()
	if err != nil {
		s.logger.Fatal(errors.Wrap(err, "save storage"))
	}

	s.logger.WithFields(logrus.Fields{
		"url":   d.OriginalURL,
		"short": d.ShortURL,
	}).Infoln("saved")
}

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
