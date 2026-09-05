// Package config
package config

import (
	"errors"
	"fmt"

	"github.com/joho/godotenv"
)

var ErrLoadingConfig = errors.New("unexpected error loading config")

type Config struct {
	AppConfig AppConfig
	DBConfig  DatabaseConfig
}

func LoadConfig(path string) (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("%w, %s", ErrLoadingConfig, err.Error())
	}

	return &Config{}, nil
}
