// Package config provides the application configuration structure and
// initialization logic using command-line flags and environment variables.
package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/wickedv43/go-shortener/internal/logger"

	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
)

// Config holds server configuration and internal logger instance.
type Config struct {
	log    *logrus.Entry // Internal logger instance scoped to the config component.
	Server Server        // Server-related configuration parameters.
}

// Server contains flags and environment-based parameters required to run the server.
type Server struct {
	FlagRunAddr     string // Address and port where the HTTP server will run.
	FlagSuffixAddr  string // Base URL used when constructing short links.
	FlagStoragePath string // Path to the JSON file used for storage recovery.
	FlagDatabaseDSN string // PostgreSQL connection string.

	FlagCertPath string // Path to certs
	FlagKeyPath  string // Path to key
	FlagHTTPS    bool   // Enable HTTPS flag

}

// Logger defines the logging level to be used in the application.
type Logger struct {
	Lvl logrus.Level // Logging level (e.g., Info, Debug).
}

// NewConfig initializes and returns a Config instance.
// It parses command-line flags, loads environment variables from a .env file,
// and applies them to override default values.
// Uses the samber/do dependency injection container.
func NewConfig(i do.Injector) (*Config, error) {
	cfg, err := do.InvokeStruct[Config](i)
	if err != nil {
		return nil, errors.Wrap(err, "invoke config")
	}

	log, err := do.Invoke[*logger.Logger](i)
	if err != nil {
		return cfg, err
	}

	cfg.log = log.WithField("component", "config")

	// flags
	flag.StringVar(&cfg.Server.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.Server.FlagSuffixAddr, "b", ":8080", "address before short url")
	flag.StringVar(&cfg.Server.FlagStoragePath, "f", "./db/storage.json", "path to database file")
	flag.StringVar(&cfg.Server.FlagDatabaseDSN, "d", "", "database connection string")
	flag.StringVar(&cfg.Server.FlagCertPath, "c", "./cert/server.crt", "path to TLS certificate file")
	flag.StringVar(&cfg.Server.FlagKeyPath, "k", "./cert/server.key", "path to TLS certificate file")
	flag.BoolVar(&cfg.Server.FlagHTTPS, "s", true, "use HTTPS")
	flag.Parse()
	err = godotenv.Load()
	if err != nil {
		cfg.log.Warn(err, "loading .env file")
	}

	// override with env vars if present
	if ServerAddr := os.Getenv("SERVER_ADDRESS"); ServerAddr != "" {
		cfg.Server.FlagRunAddr = ServerAddr
	}
	if BaseURL := os.Getenv("BASE_URL"); BaseURL != "" {
		cfg.Server.FlagSuffixAddr = BaseURL
	}
	if FileStoragePath := os.Getenv("FILE_STORAGE_PATH"); FileStoragePath != "" {
		cfg.Server.FlagStoragePath = FileStoragePath
	}
	if DatabaseDSN := os.Getenv("DATABASE_DSN"); DatabaseDSN != "" {
		cfg.Server.FlagDatabaseDSN = DatabaseDSN
	}
	if envHTTPS := os.Getenv("ENABLE_HTTPS"); envHTTPS != "" {
		cfg.Server.FlagHTTPS = envHTTPS == "true"
	}

	fmt.Println(&cfg.Server)
	return cfg, nil
}
