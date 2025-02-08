package config

import (
	"flag"
	"os"

	"github.com/joho/godotenv"
	"github.com/wickedv43/go-shortener/internal/logger"

	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Server Server
	log    *logrus.Entry
}

// Server struct
// FlagRunAddr - address and port to run server
// FlagSuffixAddr - address before short url
// FlagStoragePath - path to db recovery file
type Server struct {
	FlagRunAddr     string
	FlagSuffixAddr  string
	FlagStoragePath string
	FlagDatabaseDSN string
}

// Lvl - logs level
type Logger struct {
	Lvl logrus.Level
}

func NewConfig(i do.Injector) (*Config, error) {
	var cfg Config

	cfg.log = do.MustInvoke[*logger.Logger](i).WithField("component", "config")

	//flags
	flag.StringVar(&cfg.Server.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.Server.FlagSuffixAddr, "b", "http://localhost:8080", "address before short url")
	flag.StringVar(&cfg.Server.FlagStoragePath, "f", "./db/storage.json", "path to database file")
	flag.StringVar(&cfg.Server.FlagDatabaseDSN, "d", "", "database connection string")
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		cfg.log.Warn(err, "loading .env file")
	}

	//env
	ServerAddr := os.Getenv("SERVER_ADDRESS")
	if ServerAddr != "" {
		cfg.Server.FlagRunAddr = ServerAddr
	}

	BaseURL := os.Getenv("BASE_URL")
	if BaseURL != "" {
		cfg.Server.FlagSuffixAddr = BaseURL
	}

	FileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	if FileStoragePath != "" {
		cfg.Server.FlagStoragePath = FileStoragePath
	}

	DatabaseDSN := os.Getenv("DATABASE_DSN")
	if DatabaseDSN != "" {
		cfg.Server.FlagDatabaseDSN = DatabaseDSN
	}

	return &cfg, nil
}
