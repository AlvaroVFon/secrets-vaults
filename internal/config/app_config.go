package config

import (
	"errors"
	"fmt"
	"slices"
)

type AppConfig struct {
	Environment string `json:"env"`
}

var (
	ErrEnvironmentNotProvided = errors.New("environment must not be empty")
	ErrInvalidEnvironment     = errors.New("invalid environment provided")
)

func NewAppConfig(env string) (*AppConfig, error) {
	if env == "" {
		return nil, ErrEnvironmentNotProvided
	}

	validEnv := []string{"development", "production"}

	if !slices.Contains(validEnv, env) {
		return nil, fmt.Errorf("%w: %q, environment must be one of %v", ErrInvalidEnvironment, env, validEnv)
	}

	return &AppConfig{
		Environment: env,
	}, nil
}
