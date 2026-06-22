// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

// Package service orchestrates the collection flow: it pulls data from sources,
// persists it through repositories, and publishes signals through a Publisher.
package service

import "context"

// signalsScanChannel is the Redis pub/sub channel the engine listens on for
// "new data arrived" triggers.
const signalsScanChannel = "signals:scan"

// Publisher publishes a message to a named channel (implemented by cache.Cache).
type Publisher interface {
	Publish(ctx context.Context, channel, message string) error
}
