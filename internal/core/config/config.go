package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"github.com/playmixer/secret-keeper/internal/adapter/api/rest"
	"github.com/playmixer/secret-keeper/internal/adapter/storage"
	"github.com/playmixer/secret-keeper/internal/adapter/storage/database"
	"github.com/playmixer/secret-keeper/internal/core/uiapi"
)

// Config - конфиг сервиса.
type Config struct {
	Rest        *rest.Config
	Storage     *storage.Config
	Client      *uiapi.Config
	SecretKey   string `env:"SECRET_KEY"`
	EncryptKey  string `env:"ENCRYPT_KEY"`
	LogLevel    string `env:"LOG_LEVEL"`
	LogPath     string `env:"LOG_PATH"`
	FileMaxSize int64  `env:"FILE_MAX_SIZE"`
}

var (
	defaultFileMaxSizeUpload int64 = 819200
)

// Init - инициализация конфига.
func Init(isClient bool) (*Config, error) {
	cfg := &Config{
		Rest: &rest.Config{},
		Storage: &storage.Config{
			Database: database.Config{},
		},
		Client: &uiapi.Config{
			APIAddress: "https://localhost:8443",
		},
		FileMaxSize: defaultFileMaxSizeUpload,
	}

	cfgFile := ".env"
	if isClient {
		cfgFile = ".env.client"
	}

	err := godotenv.Load(cfgFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed load env: %w", err)
	}

	err = env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed parse environments: %w", err)
	}

	return cfg, nil
}
