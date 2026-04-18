package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"sentineldb/internal/storage"
	"sentineldb/internal/worker"
	"sentineldb/pkg/logger"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort   string
	DatabaseURL  string
	JwtSecretKey string
}

var log *logger.Logger

func loadConfig() (Config, error) {
	// Load environment variables
    godotenv.Load()

    cfg := Config{
        ServerPort:   os.Getenv("SERVER_PORT"),
        DatabaseURL:  os.Getenv("DATABASE_URL"),
        JwtSecretKey: os.Getenv("JWT_SECRET_KEY"),
    }

    if cfg.ServerPort == "" {
        return Config{}, fmt.Errorf("SERVER_PORT not set")
    }
    if cfg.DatabaseURL == "" {
        return Config{}, fmt.Errorf("DATABASE_URL not set")
    }
    if cfg.JwtSecretKey == "" {
        return Config{}, fmt.Errorf("JWT_SECRET_KEY not set")
    }

    return cfg, nil
}

func main() {
    log = logger.New(logger.Options{
		Level:  logger.LevelInfo,
		Prefix: "",
	})

	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("config error: ", err)
	}

    db, err := storage.NewConnection(cfg.DatabaseURL)
    if err != nil {
        log.Fatal("database error: ", err)
    }

    // contexto cancelado quando SIGTERM chegar
    ctx, cancel := context.WithCancel(context.Background())

    poolSize := 5
    var wg sync.WaitGroup

    for i := 0; i < poolSize; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            worker.Run(ctx, db, log)
        }()
    }

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit

    log.Info("Shutting down worker...")
    cancel()
    wg.Wait()
    log.Info("Worker stopped")
}