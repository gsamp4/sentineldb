package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sentineldb/internal/job/routes"
	"sentineldb/internal/middlewares"
	"sentineldb/internal/storage"
	"sentineldb/pkg/logger"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
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

func startServer(cfg Config, e *echo.Echo) {
	// Initialize routes, handlers, and other server setup here
	log.Info("Server starting on port " + cfg.ServerPort)
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      e,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := e.StartServer(srv); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server error: ", err)
		e.StdLogger.Panicln("Server error: ", err)
	}
}

func main() {
	// Initialize logger
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

	e := middlewares.ApplySecurityMiddlewares(echo.New())
	routes.InitRoutes(e, db, log)
	go startServer(cfg, e)

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// Notify the channel on interrupt signals
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown error: ", err)
	}
	log.Info("Server stopped")
}
