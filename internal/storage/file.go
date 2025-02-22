package storage

import (
	"bufio"
	"context"
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

func (s *FileStorage) Get(_ context.Context, url string) (Data, error) {
	file, err := s.Open()
	if err != nil {
		return Data{}, errors.Wrap(err, "open file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var d Data
		line := scanner.Bytes()

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

func (s *FileStorage) GetAll(_ context.Context, userID int) ([]Data, error) {
	file, err := s.Open()
	if err != nil {
		return []Data{}, errors.Wrap(err, "open file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	data := make([]Data, 0)

	for scanner.Scan() {
		var d Data
		line := scanner.Bytes()

		if len(line) == 0 {
			continue
		}

		if err = json.Unmarshal(line, &d); err != nil {
			return []Data{}, errors.Wrap(err, "unmarshal data")
		}

		s.log.WithField("scan", d).Info("scanning line")

		if d.UUID == userID {
			data = append(data, d)
		}
	}

	if err = scanner.Err(); err != nil {
		return []Data{}, errors.Wrap(err, "scan file")
	}

	if len(data) == 0 {
		return []Data{}, errors.New("no content")
	}

	return data, nil
}

func (s *FileStorage) BatchDelete(_ []string) error {
	return nil
}

func (s *FileStorage) HealthCheck() error {
	_, err := s.Open()
	if err != nil {
		return errors.Wrap(err, "open file")
	}
	defer s.file.Close()

	return nil
}

func (s *FileStorage) Close() error {
	//TODO implement me
	return s.file.Close()
}

func (s *FileStorage) Save(_ context.Context, d Data) error {
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

	return nil
}

// Open or Create file
func (s *FileStorage) Open() (*os.File, error) {
	return os.OpenFile(s.cfg.Server.FlagStoragePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
}

// RemoveFile remove file
func (s *FileStorage) RemoveFile() error {
	return os.Remove(s.cfg.Server.FlagStoragePath)
}
