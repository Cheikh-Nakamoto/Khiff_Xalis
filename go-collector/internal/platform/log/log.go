// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

// Package log configures the process-wide structured logger (log/slog).
package log

import (
	"log/slog"
	"os"
	"strings"
)

// Setup installs a text slog handler as the default logger. The level is read
// from the LOG_LEVEL env var (debug|info|warn|error); it defaults to info.
func Setup() *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)
	return logger
}
