package internal

import (
	"context"
	"log"
	"time"
)

// cleanupCache removes stale Redis cache keys.
//
// Keys matching "cache:*" are scanned. Keys without a TTL get one set (4h default).
// Keys already having a TTL auto-expire via Redis, so this is a safety net
// for orphaned keys that lost their TTL somehow.
func (c *Collector) CleanupCache(ctx context.Context) {
	const (
		scanCount   = 100
		defaultTTL  = 4 * time.Hour
		batchSize   = 50
	)

	iter := c.RDB.Scan(ctx, 0, "cache:*", scanCount).Iterator()
	var stale []string
	var scanned int
	var expired int

	for iter.Next(ctx) {
		key := iter.Val()
		scanned++

		// Check if key has a TTL.
		ttl, err := c.RDB.TTL(ctx, key).Result()
		if err != nil {
			log.Printf("[CACHE] TTL check error for %s: %v", key, err)
			continue
		}

		// TTL of -1 means no expiry is set — apply a default.
		if ttl == -1 {
			if err := c.RDB.Expire(ctx, key, defaultTTL).Err(); err != nil {
				log.Printf("[CACHE] Expire error for %s: %v", key, err)
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
			if err := c.RDB.Del(ctx, stale...).Err(); err != nil {
				log.Printf("[CACHE] Del batch error: %v", err)
			}
			stale = stale[:0]
		}
	}

	if err := iter.Err(); err != nil {
		log.Printf("[CACHE] Scan error: %v", err)
	}

	// Flush remaining stale keys.
	if len(stale) > 0 {
		if err := c.RDB.Del(ctx, stale...).Err(); err != nil {
			log.Printf("[CACHE] Del batch error: %v", err)
		}
	}

	log.Printf("[CACHE] Cleanup complete: scanned=%d, ttl_set=%d, stale_deleted=%d",
		scanned, expired, len(stale))
}
