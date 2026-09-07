// Package config
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var ErrLoadingConfig = errors.New("unexpected error loading config")

type Config struct {
	AppConfig AppConfig      `json:"app_config"`
	DBConfig  DatabaseConfig `json:"db_config"`
}

func LoadConfig(path string) (*Config, error) {
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("%w, %s", ErrLoadingConfig, err.Error())
	}

	// APP CONFIG
	env, err := getStringEnvVar("ENV", "")
	if err != nil {
		return nil, err
	}

	appConfig, err := NewAppConfig(env)
	if err != nil {
		return nil, err
	}

	// DB CONFIG
	dbURL, err := getStringEnvVar("POSTGRES_URL", "http://localhost:5432")
	if err != nil {
		return nil, err
	}
	dbName, err := getStringEnvVar("POSTGRES_DB", "test")
	if err != nil {
		return nil, err
	}

	dbConfig, err := NewDatabaseConfig(dbURL, dbName)
	if err != nil {
		return nil, err
	}

	return &Config{
		AppConfig: *appConfig,
		DBConfig:  *dbConfig,
	}, nil
}

func getStringEnvVar(key, defaultValue string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("%w: %q", ErrEmptyString, "key")
	}

	envVar := os.Getenv(key)
	if envVar == "" {
		return defaultValue, nil
	}

	return envVar, nil
}

func getIntEnvVar(key string) (int, error) {
	if key == "" {
		return 0, fmt.Errorf("%w: %q", ErrEmptyString, "key")
	}

	envVar := os.Getenv(key)

	intEnvVar, err := strconv.Atoi(envVar)
	if err != nil {
		return 0, err
	}

	return intEnvVar, nil
}
