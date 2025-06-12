package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestPostgresStorage_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO urls").
		WithArgs(123, "short", "original").
		WillReturnResult(sqlmock.NewResult(1, 1))

	s := &PostgresStorage{pgDB: db}
	err = s.Save(context.Background(), Data{
		UUID:        123,
		ShortURL:    "short",
		OriginalURL: "original",
	})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_Get_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"uuid", "original_url", "short_url", "is_deleted"}).
		AddRow(123, "original", "short", false)

	mock.ExpectQuery("SELECT uuid, original_url, short_url, is_deleted FROM urls").
		WithArgs("short").
		WillReturnRows(rows)

	s := &PostgresStorage{pgDB: db}
	data, err := s.Get(context.Background(), "short")
	require.NoError(t, err)
	require.Equal(t, 123, data.UUID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_Get_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("SELECT uuid, original_url, short_url, is_deleted FROM urls").
		WithArgs("notfound").
		WillReturnError(sql.ErrNoRows)

	s := &PostgresStorage{pgDB: db}
	_, err = s.Get(context.Background(), "notfound")
	require.ErrorContains(t, err, "not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_GetAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"uuid", "original_url", "short_url"}).
		AddRow(123, "http://example.com/1", "s1").
		AddRow(123, "http://example.com/2", "s2")

	mock.ExpectQuery("SELECT uuid, original_url, short_url FROM urls WHERE uuid =").
		WithArgs(123).
		WillReturnRows(rows)

	s := &PostgresStorage{pgDB: db}
	result, err := s.GetAll(context.Background(), 123)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_BatchDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"short_url"}).
		AddRow("abc123").AddRow("def456")

	mock.ExpectQuery("UPDATE urls SET is_deleted = true").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(rows)

	s := &PostgresStorage{pgDB: db}
	err = s.BatchDelete([]string{"abc123", "def456"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_BatchDelete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	empty := sqlmock.NewRows([]string{"short_url"})

	mock.ExpectQuery("UPDATE urls SET is_deleted = true").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(empty)

	s := &PostgresStorage{pgDB: db}
	err = s.BatchDelete([]string{"unknown"})
	require.ErrorContains(t, err, "not found")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_HealthCheck(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectPing()

	s := &PostgresStorage{pgDB: db}
	require.NoError(t, s.HealthCheck())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_Stats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	urlsRows := sqlmock.NewRows([]string{"count"}).
		AddRow(10)

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM urls").
		WillReturnRows(urlsRows)

	usersRows := sqlmock.NewRows([]string{"count"}).
		AddRow(5)

	mock.ExpectQuery("SELECT COUNT\\(DISTINCT uuid\\) FROM urls").
		WillReturnRows(usersRows)

	s := &PostgresStorage{pgDB: db}

	urls, users, err := s.Stats()
	require.NoError(t, err)
	require.Equal(t, 10, urls)
	require.Equal(t, 5, users)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	mock.ExpectClose()

	s := &PostgresStorage{pgDB: db}

	err = s.Close()
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresStorage_Shutdown(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	// Ожидаем, что Close будет вызван
	mock.ExpectClose()

	s := &PostgresStorage{pgDB: db, log: logrus.NewEntry(logrus.New())}

	err = s.Shutdown()
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
