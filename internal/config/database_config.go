package config

import "net/url"

type DatabaseConfig struct {
	URL    string
	DBName string
}

func NewDatabaseConfig(dbURL, dbName string) (*DatabaseConfig, error) {
	if dbURL == "" {
	}

	if _, err := url.Parse(dbURL); err != nil {
	}

	if dbName == "" {
	}

	return &DatabaseConfig{
		URL:    dbURL,
		DBName: dbName,
	}, nil
}
