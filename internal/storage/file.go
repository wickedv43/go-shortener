package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
)

type FileStorage struct {
	file *os.File
	log  *logrus.Entry
	cfg  *config.Config
}

func NewFileStorage(i do.Injector) (*FileStorage, error) {
	storage, err := do.InvokeStruct[FileStorage](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	storage.log = do.MustInvoke[*logger.Logger](i).WithField("component", "db")
	storage.cfg = do.MustInvoke[*config.Config](i)

	//create dir for file
	filePath, _ := filepath.Split(storage.cfg.Server.FlagStoragePath)
	_ = os.MkdirAll(filePath, 0755)

	// create  file
	storage.file, err = storage.Open()
	if err != nil {
		return nil, errors.Wrap(err, "create file")
	}

	defer storage.file.Close()

	return storage, err
}

func (s *FileStorage) Get(url string) (Data, error) {
	// Открываем файл
	file, err := s.Open()
	if err != nil {
		return Data{}, errors.Wrap(err, "open file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var d Data
		line := scanner.Bytes()

		// Пропускаем пустые строки
		if len(line) == 0 {
			continue
		}

		if err = json.Unmarshal(line, &d); err != nil {
			return Data{}, errors.Wrap(err, "unmarshal data")
		}

		s.log.WithField("scan", d).Info("scanning line")

		if d.ShortURL == url || d.OriginalURL == url {
			return d, nil
		}
	}

	if err = scanner.Err(); err != nil {
		return Data{}, errors.Wrap(err, "scan file")
	}

	return Data{}, errors.New("URL not found")
}

func (s *FileStorage) Delete(_ string) error {
	return nil
}

func (s *FileStorage) HealthCheck() error {
	_, err := s.Open()
	if err != nil {
		return errors.Wrap(err, "open file")
	}
	err = s.Close()
	if err != nil {
		return errors.Wrap(err, "close file")
	}
	return nil
}

func (s *FileStorage) Close() error {
	//TODO implement me
	return s.file.Close()
}

func (s *FileStorage) Save(d Data) error {
	var err error

	s.file, err = s.Open()
	if err != nil {
		return errors.Wrap(err, "open file")
	}

	defer s.file.Close()

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

func (s *FileStorage) Open() (*os.File, error) {
	return os.OpenFile(s.cfg.Server.FlagStoragePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
}
