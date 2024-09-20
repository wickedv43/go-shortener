package config

import (
	"flag"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	Server Server
	Logger Logger
}

// Server struct
// FlagRunAddr - address and port to run server
// FlagSuffixAddr - address before short url
// FlagStoragePath - path to db recovery file
type Server struct {
	FlagRunAddr     string
	FlagSuffixAddr  string
	FlagStoragePath string
}

// Lvl - logs level
type Logger struct {
	Lvl logrus.Level
}

func NewConfig(i do.Injector) (*Config, error) {
	var cfg Config
	flag.StringVar(&cfg.Server.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.Server.FlagSuffixAddr, "b", "http://localhost:8080", "address before short url")
	flag.StringVar(&cfg.Server.FlagStoragePath, "f", "./cmd/shortener/db/db", "path to database file")

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

	cfg.Logger.Lvl = logrus.InfoLevel

	flag.Parse()
	return &cfg, nil
}
