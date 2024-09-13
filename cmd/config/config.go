package config

import (
	"flag"
	"github.com/samber/do/v2"
	"github.com/sirupsen/logrus"
	"os"
)

// FlagRunAddr server ip:port
// FlagSuffixAddr address before shortered

type Config struct {
	Server Server
	Logger Logger
}

type Server struct {
	FlagRunAddr    string
	FlagSuffixAddr string
}

type Logger struct {
	Lvl logrus.Level
}

func NewConfig(i do.Injector) (*Config, error) {
	var cfg Config
	flag.StringVar(&cfg.Server.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.Server.FlagSuffixAddr, "b", "http://localhost:8080", "address before short url")

	ServerAddr := os.Getenv("SERVER_ADDRESS")
	if ServerAddr != "" {
		cfg.Server.FlagRunAddr = ServerAddr
	}

	BaseURL := os.Getenv("BASE_URL")
	if BaseURL != "" {
		cfg.Server.FlagSuffixAddr = BaseURL
	}

	cfg.Logger.Lvl = logrus.InfoLevel

	flag.Parse()
	return &cfg, nil
}
