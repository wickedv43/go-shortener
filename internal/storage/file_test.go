package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-shortener/internal/config"
)

func newTestFileStorage(t *testing.T) *FileStorage {
	t.Helper()

	dir := "testdata"
	path := filepath.Join(dir, "test_storage.json")

	_ = os.MkdirAll(dir, 0755)
	_ = os.Remove(path)

	cfg := &config.Config{
		Server: config.Server{FlagStoragePath: path},
	}
	log := logrus.New().WithField("component", "test")

	fs := &FileStorage{cfg: cfg, log: log}

	t.Cleanup(func() {
		_ = fs.Close()
		_ = os.RemoveAll(dir)
	})

	return fs
}

func TestFileStorage_SaveAndGet(t *testing.T) {
	fs := newTestFileStorage(t)

	d := Data{UUID: 1, OriginalURL: "http://original.com", ShortURL: "abc123"}
	err := fs.Save(context.Background(), d)
	require.NoError(t, err)

	got, err := fs.Get(context.Background(), "abc123")
	require.NoError(t, err)
	require.Equal(t, d.OriginalURL, got.OriginalURL)

	got2, err := fs.Get(context.Background(), "http://original.com")
	require.NoError(t, err)
	require.Equal(t, d.ShortURL, got2.ShortURL)
}

func TestFileStorage_Get_NotFound(t *testing.T) {
	fs := newTestFileStorage(t)

	_, err := fs.Get(context.Background(), "not_found")
	require.ErrorContains(t, err, "URL not found")
}

func TestFileStorage_GetAll(t *testing.T) {
	fs := newTestFileStorage(t)

	data := []Data{
		{UUID: 1, OriginalURL: "http://1.com", ShortURL: "a1"},
		{UUID: 1, OriginalURL: "http://2.com", ShortURL: "a2"},
		{UUID: 2, OriginalURL: "http://3.com", ShortURL: "a3"},
	}

	for _, d := range data {
		require.NoError(t, fs.Save(context.Background(), d))
	}

	urls, err := fs.GetAll(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, urls, 2)

	urls2, err := fs.GetAll(context.Background(), 99)
	require.ErrorContains(t, err, "no content")
	require.Len(t, urls2, 0)
}

func TestFileStorage_HealthCheck(t *testing.T) {
	fs := newTestFileStorage(t)

	err := fs.HealthCheck()
	require.NoError(t, err)
}

func TestFileStorage_Stats(t *testing.T) {
	fs := newTestFileStorage(t)

	_ = fs.Save(context.Background(), Data{UUID: 1, OriginalURL: "http://1.com", ShortURL: "a1"})
	_ = fs.Save(context.Background(), Data{UUID: 2, OriginalURL: "http://2.com", ShortURL: "a2"})

	urls, users, err := fs.Stats()

	require.NoError(t, err)
	require.GreaterOrEqual(t, urls, 2)
	require.GreaterOrEqual(t, users, 2)
}
