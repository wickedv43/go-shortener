package config

import (
	"flag"
	"github.com/samber/do/v2"
	"os"
)

// FlagRunAddr server ip:port
// FlagSuffixAddr address before shortered

type Config struct {
	FlagRunAddr    string
	FlagSuffixAddr string
}

func NewConfig(i do.Injector) (*Config, error) {
	var cfg Config
	flag.StringVar(&cfg.FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&cfg.FlagSuffixAddr, "b", "http://localhost:8080", "address before short url")

	ServerAddr := os.Getenv("SERVER_ADDRESS")
	if ServerAddr != "" {
		cfg.FlagRunAddr = ServerAddr
	}

	BaseURL := os.Getenv("BASE_URL")
	if BaseURL != "" {
		cfg.FlagSuffixAddr = BaseURL
	}

	flag.Parse()
	return &cfg, nil
}
