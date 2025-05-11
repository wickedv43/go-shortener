package storage

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"github.com/wickedv43/go-shortener/internal/config"
	"github.com/wickedv43/go-shortener/internal/logger"
)

// PostgresStorage implements the DataKeeper interface using a PostgreSQL database.
type PostgresStorage struct {
	pgDB *sql.DB        // PostgreSQL connection pool.
	log  *logrus.Entry  // Logger instance for DB-related messages.
	cfg  *config.Config // Configuration with PostgreSQL DSN.
}

// NewPostgresStorage initializes a new PostgresStorage instance using dependency injection.
// It connects to the PostgreSQL database and ensures that the "urls" table exists.
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
        original_url TEXT NOT NULL, 
        is_deleted BOOLEAN NOT NULL DEFAULT FALSE
    );`

	_, err = storage.pgDB.Exec(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create urls table")
	}

	return storage, err
}

// Save inserts a new shortened URL entry into the database.
func (s *PostgresStorage) Save(ctx context.Context, d Data) error {
	query := `INSERT INTO urls (uuid, short_url, original_url) 
          VALUES ($1, $2, $3)`

	_, err := s.pgDB.ExecContext(ctx, query, d.UUID, d.ShortURL, d.OriginalURL)
	if err != nil {
		return errors.Wrap(err, "save to postgres")
	}

	return nil
}

// Get retrieves a URL record by either its short or original form.
func (s *PostgresStorage) Get(ctx context.Context, url string) (Data, error) {
	var data Data

	query := `SELECT uuid, original_url, short_url, is_deleted FROM urls WHERE short_url = $1 OR original_url = $1`

	err := s.pgDB.QueryRowContext(ctx, query, url).Scan(&data.UUID, &data.OriginalURL, &data.ShortURL, &data.DeletedFlag)
	if errors.Is(err, sql.ErrNoRows) {
		return Data{}, errors.New("not found")
	}

	return data, nil
}

// GetAll returns all URL entries associated with the given user ID.
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

// HealthCheck checks the connection to the PostgreSQL database.
func (s *PostgresStorage) HealthCheck() error {
	return s.pgDB.Ping()
}

// BatchDelete sets the is_deleted flag to true for a batch of short URLs.
func (s *PostgresStorage) BatchDelete(short []string) error {
	query := `UPDATE urls 
          SET is_deleted = true 
          WHERE short_url = ANY($1) AND is_deleted = false
          RETURNING short_url;`

	rows, err := s.pgDB.Query(query, pq.Array(short))
	if err != nil {
		return errors.Wrap(err, "batch delete")
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
	}

	if count == 0 {
		return errors.New("not found or already deleted")
	}

	if err = rows.Err(); err != nil {
		return errors.Wrap(err, "rows iteration error")
	}

	return nil
}

// Close terminates the connection to the PostgreSQL database.
func (s *PostgresStorage) Close() error {
	return s.pgDB.Close()
}
