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

// FileStorage implements the DataKeeper interface using a plain text file.
// Each line in the file represents a JSON-encoded Data object.
type FileStorage struct {
	file *os.File       // Open file handle used for read/write operations.
	log  *logrus.Entry  // Logger instance scoped to file storage.
	cfg  *config.Config // Application configuration (contains file path).
}

// NewFileStorage initializes a FileStorage instance using dependency injection,
// ensures the file directory exists, and creates the file if it doesn't exist.
func NewFileStorage(i do.Injector) (*FileStorage, error) {
	storage, err := do.InvokeStruct[FileStorage](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	storage.log = do.MustInvoke[*logger.Logger](i).WithField("component", "db")
	storage.cfg = do.MustInvoke[*config.Config](i)

	// Ensure the directory for the file exists.
	filePath, _ := filepath.Split(storage.cfg.Server.FlagStoragePath)
	_ = os.MkdirAll(filePath, 0755)

	// Open or create the file.
	storage.file, err = storage.Open()
	if err != nil {
		return nil, errors.Wrap(err, "create file")
	}

	_ = storage.file.Close()

	return storage, err
}

// Get reads the file line by line and returns a matching Data entry by short or original URL.
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

// GetAll returns all Data records associated with a specific user ID.
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

// BatchDelete is not implemented for file storage (stub method).
func (s *FileStorage) BatchDelete(_ []string) error {
	return nil
}

// HealthCheck verifies access to the storage file by attempting to open it.
func (s *FileStorage) HealthCheck() error {
	_, err := s.Open()
	if err != nil {
		return errors.Wrap(err, "open file")
	}
	defer s.file.Close()
	return nil
}

// Close closes the currently open file handle.
func (s *FileStorage) Close() error {
	return s.file.Close()
}

// Save appends a new Data record to the file as a single JSON line.
func (s *FileStorage) Save(_ context.Context, d Data) error {
	var err error

	s.file, err = s.Open()
	if err != nil {
		return errors.Wrap(err, "open file")
	}
	defer s.file.Close()

	data, err := json.Marshal(d)
	if err != nil {
		return errors.Wrap(err, "marshal data")
	}
	data = append(data, '\n')

	_, err = s.file.Write(data)
	if err != nil {
		return errors.Wrap(err, "write data")
	}

	return nil
}

// Stats returns the number of URLs and unique users in the file storage.
func (s *FileStorage) Stats() (int, int, error) {
	file, err := s.Open()
	if err != nil {
		return 0, 0, errors.Wrap(err, "open file")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	urlCount := 0
	userSet := make(map[int]struct{})

	for scanner.Scan() {
		var d Data
		line := scanner.Bytes()

		if len(line) == 0 {
			continue
		}

		if err = json.Unmarshal(line, &d); err != nil {
			return 0, 0, errors.Wrap(err, "unmarshal data")
		}

		urlCount++
		userSet[d.UUID] = struct{}{}
	}

	if err = scanner.Err(); err != nil {
		return 0, 0, errors.Wrap(err, "scan file")
	}

	userCount := len(userSet)

	return urlCount, userCount, nil
}

// Open opens the configured file in append/read-write mode, creating it if it doesn't exist.
func (s *FileStorage) Open() (*os.File, error) {
	return os.OpenFile(s.cfg.Server.FlagStoragePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
}

// RemoveFile deletes the storage file from the filesystem.
func (s *FileStorage) RemoveFile() error {
	return os.Remove(s.cfg.Server.FlagStoragePath)
}

// Shutdown file storage
func (s *FileStorage) Shutdown() error {
	if s.file != nil {
		err := s.file.Close()
		if err != nil && !errors.Is(err, os.ErrClosed) {
			return errors.Wrap(err, "failed to close file")
		}
		s.file = nil
		s.log.Info("file storage closed")
	}
	return nil
}
