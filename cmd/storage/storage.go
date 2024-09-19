package storage

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/cmd/config"
	"github.com/wickedv43/go-shortener/cmd/logger"
	"os"
)

const dbfilename = "db.json"

type Data struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage struct {
	db     []Data
	logger *logrus.Entry
	cfg    *config.Config
}

func NewStorage(i do.Injector) (*Storage, error) {
	storage, err := do.InvokeStruct[Storage](i)
	log := do.MustInvoke[*logger.Logger](i).WithField("component", "db")
	cfg := do.MustInvoke[*config.Config](i)

	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	db := make([]Data, 0, 0)

	storage.db = db
	storage.logger = log
	storage.cfg = cfg

	return storage, err
}

func (s *Storage) Save() error {
	file, err := os.OpenFile(dbfilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return errors.Wrap(err, "open file")
	}
	defer file.Close()

	s.logger.Infof("saving data to: %s", dbfilename)

	data, err := json.MarshalIndent(s.db, "", "   ")
	if err != nil {
		return errors.Wrap(err, "marshal data")
	}
	_, err = file.Write(data)
	if err != nil {
		return errors.Wrap(err, "write data")
	}
	return nil
}

// TODO: create normal Load() func for db, repair db. :S
func (s *Storage) Load() error {
	data, err := os.ReadFile(dbfilename)
	if err != nil {
		return errors.Wrap(err, "read file")
	}

	if err = json.Unmarshal(data, &s.db); err != nil {
		return errors.Wrap(err, "unmarshal data")
	}

	fmt.Println(s.db)

	s.logger.Infof("loaded db from: %s", dbfilename)

	return nil
}
