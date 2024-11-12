package storage

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"

	_ "github.com/lib/pq"
)

type Data struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage struct {
	db      []Data
	pgDB    *sql.DB
	file    *os.File
	log     *logrus.Entry
	cfg     *config.Config
	scanner *bufio.Scanner
}

func NewStorage(i do.Injector) (*Storage, error) {
	storage, err := do.InvokeStruct[Storage](i)
	log := do.MustInvoke[*logger.Logger](i).WithField("component", "db")
	cfg := do.MustInvoke[*config.Config](i)

	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	// locMem database
	db := make([]Data, 0)

	storage.db = db
	storage.log = log
	storage.cfg = cfg

	//create dir for db file
	filePath, _ := filepath.Split(storage.cfg.Server.FlagStoragePath)
	_ = os.MkdirAll(filePath, 0755)

	// create db file
	file, err := os.OpenFile(storage.cfg.Server.FlagStoragePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, errors.Wrap(err, "create file")
	}
	storage.file = file

	// Initialize PostgreSQL connection
	pgDB, err := sql.Open("postgres", storage.cfg.Server.FlagDatabaseDSN)
	if err != nil {
		return nil, errors.Wrap(err, "connect to postgres")
	}
	storage.pgDB = pgDB

	// scanner for db file
	storage.scanner = bufio.NewScanner(storage.file)

	return storage, err
}

func (s *Storage) Ping() error {
	return s.pgDB.Ping()
}

func (s *Storage) SaveInFile(d Data) error {
	data, err := json.Marshal(d)
	data = append(data, '\n')
	if err != nil {
		return errors.Wrap(err, "marshal data")
	}

	_, err = s.file.Write(data)
	if err != nil {
		return errors.Wrap(err, "write data")
	}

	s.log.WithFields(logrus.Fields{
		"url":   d.OriginalURL,
		"short": d.ShortURL,
	}).Infof("saved to file: %s", s.cfg.Server.FlagStoragePath)

	return nil
}

// LoadFromFile() - load data from storage.json by default
func (s *Storage) LoadFromFile() error {
	dataCounter := 0

	for s.scanner.Scan() {
		var d Data

		line := s.scanner.Bytes()

		if err := json.Unmarshal(line, &d); err != nil {
			return errors.Wrap(err, "unmarshal data")
		}
		s.db = append(s.db, d)
		dataCounter++
	}

	s.log.Infof("moved %d links to locMem from: %s", dataCounter, s.cfg.Server.FlagStoragePath)

	return nil
}

func (s *Storage) saveToPostgres(d Data) error {
	query := `INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)`

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
		s.db = append(s.db, d)
		dataCounter++
	}
	if err = rows.Err(); err != nil {
		return errors.Wrap(err, "iterate postgres rows")
	}

	s.log.Infof("loaded %d links from postgres", dataCounter)
	return nil
}

func (s *Storage) Save(d Data) error {
	if s.isPostgresAvailable() {
		return s.saveToPostgres(d)
	}
	s.db = append(s.db, d)
	return s.SaveInFile(d)
}

func (s *Storage) Load() error {
	if s.isPostgresAvailable() {
		return s.loadFromPostgres()
	}

	s.log.Info("Falling back to file storage")
	return s.LoadFromFile()
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
