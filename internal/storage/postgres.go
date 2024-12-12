package storage

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *Storage) checkPostgres(url string) (string, bool, error) {
	var short string
	err := s.pgDB.QueryRow("SELECT short_url FROM urls WHERE original_url = $1", url).Scan(&short)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil
		}
		return "", false, err
	}
	return short, true, nil
}

func (s *Storage) loadFromPostgres() error {
	query := `
    CREATE TABLE IF NOT EXISTS urls (
        uuid SERIAL PRIMARY KEY,
        short_url TEXT NOT NULL,
        original_url TEXT NOT NULL
    );`

	_, err := s.pgDB.Exec(query)
	if err != nil {
		return errors.Wrap(err, "failed to create urls table")
	}

	rows, err := s.pgDB.Query(`SELECT uuid, short_url, original_url FROM urls`)
	if err != nil {
		return errors.Wrap(err, "query from postgres")
	}
	defer rows.Close()

	var dataCounter int
	for rows.Next() {
		var d Data
		if err = rows.Scan(&d.UUID, &d.ShortURL, &d.OriginalURL); err != nil {
			return errors.Wrap(err, "scan postgres row")
		}
		s.DB = append(s.DB, d)
		dataCounter++
	}
	if err = rows.Err(); err != nil {
		return errors.Wrap(err, "iterate postgres rows")
	}

	s.log.Infof("loaded %d links from postgres", dataCounter)
	return nil
}

func (s *Storage) saveToPostgres(d Data) error {
	query := `INSERT INTO urls (uuid, short_url, original_url) 
          VALUES ($1, $2, $3)`

	_, err := s.pgDB.Exec(query, d.UUID, d.ShortURL, d.OriginalURL)
	if err != nil {
		return errors.Wrap(err, "save to postgres")
	}

	s.log.WithFields(logrus.Fields{
		"url":   d.OriginalURL,
		"short": d.ShortURL,
	}).Infoln("saved to postgres")
	return nil
}

func (s *Storage) getFromPostgres(shortID string) (string, bool) {
	var originalURL string
	query := `SELECT original_url FROM urls WHERE short_url = $1`
	err := s.pgDB.QueryRow(query, shortID).Scan(&originalURL)
	if err != nil {
		return "", false
	}
	return originalURL, true
}

func (s *Storage) isPostgresAvailable() bool {
	if s.pgDB == nil {
		return false
	}
	// Attempt to ping the database to check the connection
	if err := s.pgDB.Ping(); err != nil {
		s.log.Warn("Postgres connection not available:", err)
		return false
	}

	return true
}

func (s *Storage) Ping() error {
	return s.pgDB.Ping()
}

func (s *Storage) Close() error {
	if s.file != nil {
		if err := s.file.Close(); err != nil {
			return errors.Wrap(err, "close file")
		}
	}
	if s.pgDB != nil {
		return s.pgDB.Close()
	}
	return nil
}
