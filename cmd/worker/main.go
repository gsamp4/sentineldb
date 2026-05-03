package main

import (
	"context"
	"fmt"
	stdlog "log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"sentineldb/internal/storage"
	"sentineldb/internal/worker"
	applogger "sentineldb/pkg/logger"

	"gorm.io/gorm"
)

type Config struct {
	DB  *gorm.DB
	Log *applogger.Logger
}

func LoadConfig() (*Config, error) {
	// Initialize database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	db, err := storage.NewConnection(dbURL)
	if err != nil {
		return nil, err
	}

	return &Config{
		DB:  db,
		Log: applogger.New(applogger.Options{
			Level:  applogger.LevelInfo,
			Prefix: "worker: ",
		}),
	}, nil
}

func main() {
	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		stdlog.Fatalf("Failed to load configuration: %v", err)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	w := worker.NewWorker(cfg.DB, cfg.Log)

	// GRACEFUL SHUTDOWN SETUP
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Run(ctx)
	}()

	go func(){
		<-quit
		cfg.Log.Info("Shutting down worker...")
		cancel()
	}()

	wg.Wait()
}