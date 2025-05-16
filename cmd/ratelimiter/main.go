package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Arzeeq/cloud-camp/internal/bucket"
	"github.com/Arzeeq/cloud-camp/internal/config"
	"github.com/Arzeeq/cloud-camp/internal/handler"
	"github.com/Arzeeq/cloud-camp/internal/logger"
	"github.com/Arzeeq/cloud-camp/internal/ratelimiter"
	"github.com/Arzeeq/cloud-camp/internal/service"
	"github.com/Arzeeq/cloud-camp/internal/storage/pg"
	"github.com/jackc/pgx/v5/pgxpool"
)

type myHandler int

func (h *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
}

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// parse command line arguments
	configName := flag.String("config", "ratelimiter.yaml", "config filename for loadbalancer")
	logFormat := flag.String("log-format", "text", "set on of [text, json] logger format")
	logLevel := flag.String("log-level", "info", "include [debug, info, warn, error] logs. Each logging level automatically includes stricter levels")

	flag.Parse()
	if configName == nil || logFormat == nil || logLevel == nil {
		log.Fatal("required command line arguments were not provided")
	}

	// initialize logger
	l := logger.New(*logFormat, *logLevel)
	l.Info("logger was initialized", slog.String("format", *logFormat), slog.String("level", *logLevel))

	// load config
	cfg, err := config.LoadConfigRateLimiter(*configName)
	if err != nil {
		l.Error(err.Error())
		return
	}
	l.Info("config was loaded")

	// initialize token service
	tokenService, err := initTokenService(cfg.MigrationDir, cfg.GetConnStr())
	if err != nil {
		l.Error(err.Error())
		return
	}
	l.Info("token service was initialized")

	// mount token handler
	tokenHandler := handler.NewTokenHandler(tokenService, l)
	go func() {
		l.Info("starting listening", slog.Int("port", cfg.TokenPort))
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.TokenPort), http.HandlerFunc(tokenHandler.SetCapacity)); err != nil {
			l.Error("token handler has encountered an error")
		}
	}()

	b := bucket.New(cfg.DefaultCapacity, cfg.Interval, tokenService)
	defer b.Stop()

	go func() {
		var h myHandler
		l.Info("starting listening", slog.Int("port", cfg.Port))
		http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), ratelimiter.New(b, l).Middleware(&h))
	}()

	<-quit
	l.Info("Gracefully shutting down application")
}

func initTokenService(migDir, connStr string) (*service.TokenService, error) {
	// initialize connections pool
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connections pool: %v", err.Error())
	}

	// migrate up
	migrator := pg.NewMigrator(migDir, connStr)
	if err := migrator.Up(); err != nil {
		return nil, fmt.Errorf("failed to migrate up: %v", err.Error())
	}

	tokenStorage := pg.NewTokenStorage(pool)
	return service.NewTokenService(tokenStorage), nil
}
