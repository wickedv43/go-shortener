package config

import (
	"flag"
	"os"
	"testing"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
	"github.com/wickedv43/go-shortener/internal/logger"
)

func resetConfig() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	_ = os.Unsetenv("SERVER_ADDRESS")
	_ = os.Unsetenv("BASE_URL")
	_ = os.Unsetenv("FILE_STORAGE_PATH")
	_ = os.Unsetenv("DATABASE_DSN")
}

func TestNewConfig_DefaultFlagsAndEnv(t *testing.T) {
	resetConfig()

	os.Args = []string{"cmd"}

	os.Setenv("SERVER_ADDRESS", ":9999")
	os.Setenv("BASE_URL", "https://short.io")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/test.json")
	os.Setenv("DATABASE_DSN", "postgres://test")

	container := do.New()
	do.Provide(container, logger.NewLogger)

	cfg, err := NewConfig(container)
	require.NoError(t, err)

	require.Equal(t, ":9999", cfg.Server.FlagRunAddr)
	require.Equal(t, "https://short.io", cfg.Server.FlagSuffixAddr)
	require.Equal(t, "/tmp/test.json", cfg.Server.FlagStoragePath)
	require.Equal(t, "postgres://test", cfg.Server.FlagDatabaseDSN)
}

func TestNewConfig_CustomFlags(t *testing.T) {
	resetConfig()

	os.Args = []string{
		"cmd",
		"-a=:9990",
		"-b=https://short.test",
		"-f=/tmp/db.json",
		"-d=dsn_from_flag",
	}

	container := do.New()
	do.Provide(container, logger.NewLogger)
	do.Provide(container, func(i do.Injector) (*Config, error) {
		return &Config{}, nil
	})

	cfg, err := NewConfig(container)
	require.NoError(t, err)
	require.Equal(t, ":9990", cfg.Server.FlagRunAddr)
	require.Equal(t, "https://short.test", cfg.Server.FlagSuffixAddr)
	require.Equal(t, "/tmp/db.json", cfg.Server.FlagStoragePath)
	require.Equal(t, "dsn_from_flag", cfg.Server.FlagDatabaseDSN)
}

func TestNewConfig_InvokeLoggerError(t *testing.T) {
	resetConfig()

	container := do.New()
	do.Provide(container, func(i do.Injector) (*Config, error) {
		return &Config{}, nil
	})

	_, err := NewConfig(container)
	require.Error(t, err)
}
