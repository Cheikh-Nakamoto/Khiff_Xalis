// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

// Package cache wraps Redis for pub/sub signalling and cache housekeeping.
package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is a thin adapter over a Redis client.
type Cache struct {
	rdb *redis.Client
}

// New creates a Cache.
func New(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

// Publish posts a message to a Redis pub/sub channel.
func (c *Cache) Publish(ctx context.Context, channel, message string) error {
	return c.rdb.Publish(ctx, channel, message).Err()
}

// Cleanup removes stale Redis cache keys.
//
// Keys matching "cache:*" are scanned. Keys without a TTL get one set (4h default).
// Keys already having a TTL auto-expire via Redis, so this is a safety net
// for orphaned keys that lost their TTL somehow.
func (c *Cache) Cleanup(ctx context.Context) {
	const (
		scanCount  = 100
		defaultTTL = 4 * time.Hour
		batchSize  = 50
	)

	iter := c.rdb.Scan(ctx, 0, "cache:*", scanCount).Iterator()
	var stale []string
	var scanned int
	var expired int
	var deleted int

	flush := func() {
		if len(stale) == 0 {
			return
		}
		if err := c.rdb.Del(ctx, stale...).Err(); err != nil {
			slog.Error("cache del batch failed", "err", err)
		} else {
			deleted += len(stale)
		}
		stale = stale[:0]
	}

	for iter.Next(ctx) {
		key := iter.Val()
		scanned++

		// Check if key has a TTL.
		ttl, err := c.rdb.TTL(ctx, key).Result()
		if err != nil {
			slog.Error("cache ttl check failed", "key", key, "err", err)
			continue
		}

		// TTL of -1 means no expiry is set — apply a default.
		if ttl == -1 {
			if err := c.rdb.Expire(ctx, key, defaultTTL).Err(); err != nil {
				slog.Error("cache expire failed", "key", key, "err", err)
			} else {
				expired++
			}
		}

		// TTL of -2 means key does not exist (already expired between scan and check).
		if ttl == -2 {
			stale = append(stale, key)
		}

		// Batch-delete keys that are already gone (housekeeping).
		if len(stale) >= batchSize {
			flush()
		}
	}

	if err := iter.Err(); err != nil {
		slog.Error("cache scan failed", "err", err)
	}

	// Flush remaining stale keys.
	flush()

	slog.Info("cache cleanup complete", "scanned", scanned, "ttl_set", expired, "stale_deleted", deleted)
}
