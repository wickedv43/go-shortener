// Package config provides the application configuration structure and
// initialization logic using command-line flags and environment variables.
package config

import (
	"encoding/json"
	"flag"
	"os"
	"strings"

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
	FlagRunAddr     string `json:"server_address"`    // Address and port where the HTTP server will run.
	FlagSuffixAddr  string `json:"base_url"`          // Base URL used when constructing short links.
	FlagStoragePath string `json:"file_storage_path"` // Path to the JSON file used for storage recovery.
	FlagDatabaseDSN string `json:"database_dsn"`      // Postgres connection string.

	FlagCertPath string `json:"server_crt_path"` // Path to certs
	FlagKeyPath  string `json:"server_ket_path"` // Path to key (NB: используем "ket" как в JSON)
	FlagHTTPS    bool   `json:"enable_https"`    // Enable HTTPS flag

	FlagTrustedSubnet string `json:"trusted_subnet"` //Trusted subnet

	FlagConfigPath string `json:"-"` // Path to JSON config file (не парсится из JSON)
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
	flag.StringVar(&cfg.Server.FlagConfigPath, "c", "", "path to JSON config file")
	flag.StringVar(&cfg.Server.FlagConfigPath, "config", "", "path to JSON config file")
	flag.StringVar(&cfg.Server.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.Server.FlagSuffixAddr, "b", "http://localhost:8080", "address before short url")
	flag.StringVar(&cfg.Server.FlagStoragePath, "f", "./db/storage.json", "path to database file")
	flag.StringVar(&cfg.Server.FlagDatabaseDSN, "d", "", "database connection string")
	flag.StringVar(&cfg.Server.FlagCertPath, "sc", "./cert/server.crt", "path to TLS certificate file")
	flag.StringVar(&cfg.Server.FlagKeyPath, "sk", "./cert/server.key", "path to TLS certificate file")
	flag.BoolVar(&cfg.Server.FlagHTTPS, "s", false, "use HTTPS")
	flag.StringVar(&cfg.Server.FlagTrustedSubnet, "t", "", "trusted subnet")

	// override with env vars if present
	err = godotenv.Load()
	if err != nil {
		cfg.log.Warn(err, "loading .env file")
	}

	if cfg.Server.FlagConfigPath == "" {
		cfg.Server.FlagConfigPath = os.Getenv("CONFIG")
	}

	cfg.loadFromJSON()

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

	if TrustedSubnet := os.Getenv("TRUSTED_SUBNET"); TrustedSubnet != "" {
		cfg.Server.FlagTrustedSubnet = TrustedSubnet
	}

	if cfg.Server.FlagSuffixAddr != "" {
		if cfg.Server.FlagHTTPS {
			if strings.HasPrefix(cfg.Server.FlagSuffixAddr, "http://") {
				cfg.Server.FlagSuffixAddr = "https://" + strings.TrimPrefix(cfg.Server.FlagSuffixAddr, "http://")
			}
		}
	}

	if !flag.Parsed() {
		flag.Parse()
	}

	return cfg, nil
}

// Parsing json cfg
func (c *Config) loadFromJSON() {
	if c.Server.FlagConfigPath == "" {
		return
	}

	config, err := os.ReadFile(c.Server.FlagConfigPath)
	if err != nil {
		c.log.Warnf("error opening config file: %v", err)
	}

	err = json.Unmarshal(config, &c.Server)
	if err != nil {
		c.log.Warnf("error parsing config file: %v", err)
	}
}
