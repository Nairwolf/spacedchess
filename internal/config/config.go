// Package config loads runtime configuration from the environment.
package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL string
	Addr        string
}

func Load() (Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		DatabaseURL: databaseURL,
		Addr:        ":" + port,
	}, nil
}
