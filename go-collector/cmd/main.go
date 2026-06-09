// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/brvm/go-collector/internal"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// BCEAO macroeconomic data: 1st of each month at 8h GMT.
	cronScheduler.AddFunc("0 8 1 * *", func() {
		log.Println("[MAIN] Running BCEAO macroeconomic collection...")
		if err := c.CollectBCEAO(ctx); err != nil {
			log.Printf("[MAIN] BCEAO error: %v", err)
		}
	})

	// Commodity prices (cocoa, oil): daily at 17h GMT (after markets close).
	cronScheduler.AddFunc("0 17 * * *", func() {
		log.Println("[MAIN] Running commodity price collection...")
		if err := c.CollectCommodities(ctx); err != nil {
			log.Printf("[MAIN] Commodity error: %v", err)
		}
	})

	// Political risk (ECOWAS, GDELT): daily at 7h GMT (before market opens).
	cronScheduler.AddFunc("0 7 * * *", func() {
		log.Println("[MAIN] Running ECOWAS political risk collection...")
		if err := c.CollectPoliticalRisk(ctx); err != nil {
			log.Printf("[MAIN] Political risk error: %v", err)
		}
	})

	cronScheduler.Start()
	log.Println("[MAIN] Scheduler started. Waiting for cron jobs...")

	// Start prometheus metrics server on 9090
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("[MAIN] Starting metrics server on :9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Printf("[MAIN] Metrics server error: %v", err)
		}
	}()

	// Block forever.
	select {}
}
