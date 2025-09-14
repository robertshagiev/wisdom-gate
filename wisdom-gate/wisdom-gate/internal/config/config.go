package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server ServerConfig
	Redis  RedisConfig
	POW    POWConfig
	Quotes QuotesConfig
	Repo   struct {
		ConnectionString string `envconfig:"DBSTRING" required:"true"`
		MigrationPath    string `envconfig:"MIGRATION_PATH" default:"/opt/migrations"`
	}
}

type ServerConfig struct {
	Port         string        `envconfig:"SERVER_PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT" default:"30s"`
	WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`
	IdleTimeout  time.Duration `envconfig:"IDLE_TIMEOUT" default:"60s"`
	MaxConns     int           `envconfig:"MAX_CONNECTIONS" default:"100"`
	RateLimit    int           `envconfig:"RATE_LIMIT" default:"10"`
	RateWindow   time.Duration `envconfig:"RATE_WINDOW" default:"1m"`
}

type RedisConfig struct {
	Addr         string        `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	Password     string        `envconfig:"REDIS_PASSWORD" default:""`
	DB           int           `envconfig:"REDIS_DB" default:"0"`
	ChallengeTTL time.Duration `envconfig:"CHALLENGE_TTL" default:"20s"`
	SpentTTL     time.Duration `envconfig:"SPENT_TTL" default:"2m"`
}

type POWConfig struct {
	Difficulty int `envconfig:"POW_DIFFICULTY" default:"20"`
}

type QuotesConfig struct {
	Source string `envconfig:"QUOTES_SOURCE" default:"internal"`
}

func NewConfig() (*Config, error) {
	var config Config

	if err := envconfig.Process("", &config.Server); err != nil {
		return nil, fmt.Errorf("failed to parse wisdom-gate config: %w", err)
	}

	if err := envconfig.Process("", &config.Redis); err != nil {
		return nil, fmt.Errorf("failed to parse redis config: %w", err)
	}

	if err := envconfig.Process("", &config.POW); err != nil {
		return nil, fmt.Errorf("failed to parse pow config: %w", err)
	}

	if err := envconfig.Process("", &config.Quotes); err != nil {
		return nil, fmt.Errorf("failed to parse quotes config: %w", err)
	}

	if err := envconfig.Process("", &config.Repo); err != nil {
		return nil, fmt.Errorf("failed to parse repo config: %w", err)
	}

	return &config, nil
}
