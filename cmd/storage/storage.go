package storage

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/cmd/config"
	"github.com/wickedv43/go-shortener/cmd/logger"
	"os"
	"path/filepath"
	"strings"
)

type Data struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Storage struct {
	db  []Data
	log *logrus.Entry
	cfg *config.Config
}

func NewStorage(i do.Injector) (*Storage, error) {
	storage, err := do.InvokeStruct[Storage](i)
	log := do.MustInvoke[*logger.Logger](i).WithField("component", "db")
	cfg := do.MustInvoke[*config.Config](i)

	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	db := make([]Data, 0)

	storage.db = db
	storage.log = log
	storage.cfg = cfg

	return storage, err
}

func (s *Storage) Save(d Data) error {

	filePath, _ := filepath.Split(s.cfg.Server.FlagStoragePath)

	_ = os.MkdirAll(filePath, 0755)

	file, err := os.OpenFile(s.cfg.Server.FlagStoragePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		return errors.Wrap(err, "open file")
	}
	defer file.Close()

	s.log.Infof("saving data to: %s", s.cfg.Server.FlagStoragePath)

	data, err := json.Marshal(d)
	if err != nil {
		return errors.Wrap(err, "marshal data")
	}
	_, err = file.Write(data)
	if err != nil {
		return errors.Wrap(err, "write data")
	}

	_, err = file.Write([]byte("\n"))
	if err != nil {
		return errors.Wrap(err, "write data")
	}

	return nil
}

// Load() - load data from file.json by default
func (s *Storage) Load() error {
	var bd []byte

	filePath, _ := filepath.Split(s.cfg.Server.FlagStoragePath)

	_ = os.MkdirAll(filePath, 0755)

	file, err := os.OpenFile(s.cfg.Server.FlagStoragePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return errors.Wrap(err, "open file")
	}
	defer file.Close()

	_, err = file.Read(bd)
	if err != nil {
		return errors.Wrap(err, "read file")
	}

	data := strings.Split(string(bd), "\n")

	for _, sd := range data {
		var d Data
		if err = json.Unmarshal([]byte(sd), &d); err != nil {
			return errors.Wrap(err, "unmarshal data")
		}
		s.db = append(s.db, d)
	}

	return nil
}
