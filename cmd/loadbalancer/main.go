package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Arzeeq/cloud-camp/internal/config"
	"github.com/Arzeeq/cloud-camp/internal/healthcheck"
	"github.com/Arzeeq/cloud-camp/internal/loadbalancer"
	"github.com/Arzeeq/cloud-camp/internal/logger"
	"github.com/Arzeeq/cloud-camp/internal/pool"
)

func main() {
	configName := flag.String("config", "config.yaml", "config filename for loadbalancer")
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
	cfg, err := config.LoadConfig(*configName)
	if err != nil {
		l.Error(err.Error())
		return
	}
	l.Info("config was loaded")

	// initialize servers pool
	pool, err := initPool(cfg.Algorithm, cfg.Servers)
	if err != nil {
		l.Error(err.Error())
		return
	}
	l.Info("servers pool initialized")

	// initialize health checker
	hc := healthcheck.New(pool, l, cfg.HealthCheckInterval)
	hc.Start()
	defer hc.Stop()
	l.Info("healthchecker was activated")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		lb, err := loadbalancer.New(pool, l)
		if err != nil {
			l.Error(fmt.Sprintf("failed to create load balancer instance: %v", err))
			return
		}

		l.Info("starting load balancer", slog.Int("port", cfg.Port))
		if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), lb); err != nil {
			l.Error(fmt.Sprintf("load balancer has encountered an error: %v", err))
		}
	}()

	<-stop
	l.Info("Shutting down load balancer gracefully")
}

func initPool(alg pool.Algo, servers []string) (pool.Pooler, error) {
	switch alg {
	case pool.RoundRobin:
		return pool.NewRoundRobinPool(servers)
	}

	return nil, errors.New("unexpected algorith name")
}
