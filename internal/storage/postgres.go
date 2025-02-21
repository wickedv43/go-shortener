package storage

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
)

type PostgresStorage struct {
	pgDB *sql.DB
	log  *logrus.Entry
	cfg  *config.Config
}

func NewPostgresStorage(i do.Injector) (*PostgresStorage, error) {
	storage, err := do.InvokeStruct[PostgresStorage](i)
	log := do.MustInvoke[*logger.Logger](i).WithField("component", "db")
	cfg := do.MustInvoke[*config.Config](i)

	if err != nil {
		return nil, errors.Wrap(err, "invoke struct")
	}

	storage.log = log
	storage.cfg = cfg

	pgDB, err := sql.Open("postgres", storage.cfg.Server.FlagDatabaseDSN)
	if err != nil {
		return nil, errors.Wrap(err, "connect to postgres")
	}
	storage.pgDB = pgDB

	query := `
    CREATE TABLE IF NOT EXISTS urls (
        uuid SERIAL NOT NULL,
        short_url TEXT NOT NULL,
        original_url TEXT NOT NULL,  -- ✅ Исправлено
        is_deleted BOOLEAN NOT NULL DEFAULT FALSE  -- ✅ Исправлено
    );`

	_, err = storage.pgDB.Exec(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create urls table")
	}

	return storage, err
}

func (s *PostgresStorage) Save(ctx context.Context, d Data) error {
	query := `INSERT INTO urls (uuid, short_url, original_url) 
          VALUES ($1, $2, $3)`

	_, err := s.pgDB.ExecContext(ctx, query, d.UUID, d.ShortURL, d.OriginalURL)
	if err != nil {
		return errors.Wrap(err, "save to postgres")
	}

	return nil
}

func (s *PostgresStorage) Get(ctx context.Context, url string) (Data, error) {
	var data Data

	query := `SELECT uuid, original_url, short_url FROM urls WHERE short_url = $1 OR original_url = $1`

	err := s.pgDB.QueryRowContext(ctx, query, url).Scan(&data.UUID, &data.OriginalURL, &data.ShortURL)
	if errors.Is(err, sql.ErrNoRows) {
		return Data{}, errors.New("not found")
	}

	return data, nil
}

func (s *PostgresStorage) GetAll(ctx context.Context, userID int) ([]Data, error) {
	var data []Data
	var err error

	query := `SELECT uuid, original_url, short_url FROM urls WHERE uuid = $1`

	rows, err := s.pgDB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "get all urls query failed")
	}
	defer rows.Close()

	for rows.Next() {
		var d Data
		if err = rows.Scan(&d.UUID, &d.OriginalURL, &d.ShortURL); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}
		data = append(data, d)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows iteration error")
	}

	return data, nil
}

func (s *PostgresStorage) HealthCheck() error {
	return s.pgDB.Ping()
}

func (s *PostgresStorage) BatchDelete(short string) error {
	tx, err := s.pgDB.Begin()
	if err != nil {
		return err
	}

	query := `SELECT short_url FROM urls WHERE short_url = $1 FOR UPDATE;`
	err = tx.QueryRow(query, short).Scan(&short)
	if err != nil {
		tx.Rollback()
		return err
	}

	updateQuery := `UPDATE urls 
	                SET is_deleted = true 
	                WHERE short_url = $1 AND is_deleted = false;`

	result, err := tx.Exec(updateQuery, short)
	if err != nil {
		tx.Rollback()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return errors.New("not found")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	s.log.Infof("Deleted %d urls", rowsAffected)
	return nil
}

func (s *PostgresStorage) Close() error {
	return s.pgDB.Close()
}
