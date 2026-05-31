// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package main

import (
	"context"
	"log"
	"time"

	"github.com/brvm/go-collector/internal"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
)

func main() {
	ctx := context.Background()
	cfg := internal.LoadConfig()

	// Connect PostgreSQL.
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Connect Redis.
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer rdb.Close()

	c := internal.NewCollector(db, rdb, cfg)

	// Scheduler — all times in Africa/Abidjan (GMT+0).
	cronScheduler := cron.New(cron.WithLocation(time.FixedZone("GMT", 0)))

	// GitHub CSV: every 15 min, Mon-Fri 9h-15h GMT (BRVM trading hours).
	cronScheduler.AddFunc("*/15 9-15 * * 1-5", func() {
		log.Println("[MAIN] Running GitHub CSV collection...")
		if err := c.CollectGithubCSV(ctx); err != nil {
			log.Printf("[MAIN] GitHub CSV error: %v", err)
		}
	})

	// Sikafinance: daily at 18h GMT (after market close).
	cronScheduler.AddFunc("0 18 * * *", func() {
		log.Println("[MAIN] Running Sikafinance collection...")
		if err := c.CollectSikafinance(ctx); err != nil {
			log.Printf("[MAIN] Sikafinance error: %v", err)
		}
	})

	// FluxBourse: every 30 min, 9h-15h GMT.
	cronScheduler.AddFunc("*/30 9-15 * * *", func() {
		log.Println("[MAIN] Running FluxBourse collection...")
		if err := c.CollectFluxBourse(ctx); err != nil {
			log.Printf("[MAIN] FluxBourse error: %v", err)
		}
	})

	// Cache cleanup: daily at 2h GMT.
	cronScheduler.AddFunc("0 2 * * *", func() {
		log.Println("[MAIN] Running cache cleanup...")
		c.CleanupCache(ctx)
	})

	cronScheduler.Start()
	log.Println("[MAIN] Scheduler started. Waiting for cron jobs...")

	// Block forever.
	select {}
}
