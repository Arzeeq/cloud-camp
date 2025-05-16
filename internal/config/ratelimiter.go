package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type DBParam struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}

func (p *DBParam) GetConnStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", p.DBUser, p.DBPassword, p.DBHost, p.DBPort, p.DBName)
}

type RateLimiter struct {
	Port            int           `yaml:"port"`
	TokenPort       int           `yaml:"token_port"`
	MigrationDir    string        `yaml:"migration_dir"`
	Interval        time.Duration `yaml:"interval"`
	DefaultCapacity int           `yaml:"default_capacity"`
	DBParam         `yaml:"-"`
}

func LoadConfigRateLimiter(filename string) (*RateLimiter, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg RateLimiter
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.DBPassword = os.Getenv("DATABASE_PASSWORD")
	cfg.DBUser = os.Getenv("DATABASE_USER")
	cfg.DBHost = os.Getenv("DATABASE_HOST")
	cfg.DBPort = os.Getenv("DATABASE_PORT")
	cfg.DBName = os.Getenv("DATABASE_NAME")

	return &cfg, nil
}
