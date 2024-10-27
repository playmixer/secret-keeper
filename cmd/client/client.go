package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/playmixer/secret-keeper/internal/adapter/logger"
	"github.com/playmixer/secret-keeper/internal/adapter/storage/file"
	"github.com/playmixer/secret-keeper/internal/adapter/ui"
	"github.com/playmixer/secret-keeper/internal/core/config"
	"github.com/playmixer/secret-keeper/internal/core/uiapi"
)

var (
	shutdownDelay = time.Second * 2

	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Println("Build verson: " + buildVersion)
	fmt.Println("Build date: " + buildDate)
	fmt.Println("Build commit: " + buildCommit)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Init(true)
	if err != nil {
		return fmt.Errorf("failed initialize config: %w", err)
	}

	lgr, err := logger.New(
		logger.SetLevel(cfg.LogLevel),
		logger.SetLogPath(cfg.LogPath),
		logger.SetEnableTerminalOutput(false),
	)
	if err != nil {
		return fmt.Errorf("failed initialize logger: %w", err)
	}
	defer func() {
		err := lgr.Sync()
		if err != nil {
			lgr.Error("failed sync logger", zap.Error(err))
		}
	}()

	store, err := file.Init(file.SetLogger(lgr))
	if err != nil {
		return fmt.Errorf("failed initialize store: %w", err)
	}

	api, err := uiapi.New(
		ctx,
		store,
		lgr,
		uiapi.SetConfig(*cfg.Client),
		uiapi.SetFileMaxSize(cfg.FileMaxSize),
	)
	if err != nil {
		return fmt.Errorf("failed create client api: %w", err)
	}

	client, err := ui.New(ctx, api, lgr, ui.SetVersion(buildVersion, buildDate, buildCommit))
	if err != nil {
		return fmt.Errorf("failed initialize client: %w", err)
	}
	go func() {
		defer cancel()
		if err := client.Run(nil); err != nil {
			lgr.Error("failed run client", zap.Error(err))
		}
	}()
	<-ctx.Done()
	lgr.Info("Stopping...")
	ctxShutdown, stop := context.WithTimeout(context.Background(), shutdownDelay)
	defer stop()

	client.Close()
	<-ctxShutdown.Done()
	lgr.Info("Client closed")

	return nil
}
