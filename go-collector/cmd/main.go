// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brvm/go-collector/internal"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
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
	cronScheduler := internal.NewCronScheduler(c, ctx)
	
	cronScheduler.Start()
	log.Println("[MAIN] Scheduler started. Waiting for cron jobs...")

	http.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Addr: ":9090"}

	go func() {
		log.Println("[MAIN] Starting metrics server on :9090")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[MAIN] Metrics server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[MAIN] Shutting down metrics server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[MAIN] Server forced to shutdown: %v", err)
	}
	log.Println("[MAIN] Server exited")
}
