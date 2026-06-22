// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

// Package config loads collector configuration from the environment.
package config

import "os"

// Config holds all environment-sourced configuration.
type Config struct {
	DatabaseURL    string
	RedisURL       string
	GithubCSVURL   string
	SikafinanceURL string
	FluxbourseURL  string
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		RedisURL:       os.Getenv("REDIS_URL"),
		GithubCSVURL:   os.Getenv("GITHUB_CSV_URL"),
		SikafinanceURL: os.Getenv("SIKAFINANCE_URL"),
		FluxbourseURL:  os.Getenv("FLUXBOURSE_URL"),
	}
}
