package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
)

type Data struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage struct {
	db      []Data
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

	// scanner for db file
	storage.scanner = bufio.NewScanner(storage.file)

	return storage, err
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

	s.log.Infof("saved to file: %s", s.cfg.Server.FlagStoragePath)

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

func (s *Storage) Close() error {
	return s.file.Close()
}
