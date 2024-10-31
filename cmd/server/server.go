package main

import (
	"fmt"
	"log"

	"go.uber.org/zap"

	"github.com/playmixer/secret-keeper/internal/adapter/api/rest"
	"github.com/playmixer/secret-keeper/internal/adapter/logger"
	"github.com/playmixer/secret-keeper/internal/adapter/storage/database"
	"github.com/playmixer/secret-keeper/internal/core/config"
	"github.com/playmixer/secret-keeper/internal/core/keeper"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Init(false)
	if err != nil {
		return fmt.Errorf("failed initialize config: %w", err)
	}

	lgr, err := logger.New(logger.SetLevel(cfg.LogLevel), logger.SetLogPath(cfg.LogPath))
	if err != nil {
		return fmt.Errorf("failed inittialize logger: %w", err)
	}
	defer func() {
		err := lgr.Sync()
		if err != nil {
			lgr.Error("failed sync logger", zap.Error(err))
		}
	}()

	store, err := database.New(cfg.Storage.Database.DSN)
	if err != nil {
		return fmt.Errorf("failed initialize storage: %w", err)
	}

	keep, err := keeper.New(store, keeper.SetEncryptKey(cfg.EncryptKey))
	if err != nil {
		return fmt.Errorf("failed initialize keeper: %w", err)
	}

	srv, err := rest.New(
		keep,
		rest.SetConfig(*cfg.Rest),
		rest.SetSecretKey(cfg.SecretKey),
		rest.SetLogger(lgr),
		rest.SetSSLEnable(cfg.Rest.SSLEnable),
	)
	if err != nil {
		return fmt.Errorf("failed initialize server: %w", err)
	}

	if err := srv.Run(); err != nil {
		return fmt.Errorf("failed run server: %w", err)
	}

	return nil
}
