package config

import (
	"flag"
	"os"
)

// FlagRunAddr server ip:port
// FlagSuffixAddr address before shortered
var (
	FlagRunAddr    string
	FlagSuffixAddr string
)

// Parse arguments from env or flags
func ParseFlags() {
	flag.StringVar(&FlagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&FlagSuffixAddr, "b", "http://localhost:8080", "address before short url")

	ServerAddr := os.Getenv("SERVER_ADDRESS")
	if ServerAddr != "" {
		FlagRunAddr = ServerAddr
	}

	BaseURL := os.Getenv("BASE_URL")
	if BaseURL != "" {
		FlagSuffixAddr = BaseURL
	}

	flag.Parse()
}
