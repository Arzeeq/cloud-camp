package config

import (
	"fmt"
	"os"
	"time"

	"github.com/Arzeeq/cloud-camp/internal/pool"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Port                int           `yaml:"port"`
	Algorithm           pool.Algo     `yaml:"algorithm"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval"`
	Servers             []string      `yaml:"servers"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.HealthCheckInterval == 0 {
		cfg.HealthCheckInterval = 10 * time.Second
	}
	if cfg.Algorithm == pool.Undefined {
		cfg.Algorithm = pool.RoundRobin
	}

	return &cfg, nil
}
