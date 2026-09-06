package config

import (
	"errors"
	"fmt"
	"net/url"
)

type DatabaseConfig struct {
	URL    string `json:"url"`
	DBName string `json:"dbName"`
}

var (
	ErrEmptyString       = errors.New("invalid empty argument provided")
	ErrInvalidConnString = errors.New("invalid connection string provided")
)

func NewDatabaseConfig(dbURL, dbName string) (*DatabaseConfig, error) {
	if dbURL == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyString, "dbURL")
	}

	if _, err := url.Parse(dbURL); err != nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidConnString, dbURL)
	}

	if dbName == "" {
		return nil, fmt.Errorf("%w: %q", ErrEmptyString, "dbName")
	}

	return &DatabaseConfig{
		URL:    dbURL,
		DBName: dbName,
	}, nil
}
