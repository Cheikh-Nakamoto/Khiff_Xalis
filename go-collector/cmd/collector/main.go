// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brvm/go-collector/internal/cache"
	"github.com/brvm/go-collector/internal/config"
	platformlog "github.com/brvm/go-collector/internal/platform/log"
	"github.com/brvm/go-collector/internal/repository/postgres"
	"github.com/brvm/go-collector/internal/scheduler"
	"github.com/brvm/go-collector/internal/service"
	"github.com/brvm/go-collector/internal/source"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func main() {
	platformlog.Setup()

	ctx := context.Background()
	cfg := config.Load()

	// Connect PostgreSQL.
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	// Connect Redis.
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()

	// Infrastructure adapters.
	httpClient := source.NewHTTPClient()
	redisCache := cache.New(rdb)

	// Repositories.
	marketRepo := postgres.NewMarketRepo(db)
	fundamentalRepo := postgres.NewFundamentalRepo(db)
	macroRepo := postgres.NewMacroRepo(db)

	// Sources.
	githubSrc := source.NewGithubSource(cfg.GithubCSVURL, httpClient)
	sikaSrc := source.NewSikafinanceSource(cfg.SikafinanceURL, httpClient)
	fluxSrc := source.NewFluxBourseSource(cfg.FluxbourseURL)
	bceaoSrc := source.NewBCEAOSource(httpClient)
	commoditySrc := source.NewCommoditySource(httpClient)
	politicalSrc := source.NewPoliticalRiskSource(httpClient)

	// Services.
	marketSvc := service.NewMarketService(githubSrc, sikaSrc, marketRepo, redisCache)
	fundamentalSvc := service.NewFundamentalService(fluxSrc, fundamentalRepo, redisCache)
	macroSvc := service.NewMacroService(bceaoSrc, commoditySrc, politicalSrc, macroRepo)

	// Scheduler — all times in Africa/Abidjan (GMT+0).
	jobs := []scheduler.Job{
		// GitHub CSV: every 15 min, Mon-Fri 9h-15h GMT (BRVM trading hours).
		{Name: "github_csv", Spec: "*/15 9-15 * * 1-5", Run: scheduler.Wrap(ctx, "github_csv", marketSvc.CollectGithubCSV)},
		// Sikafinance: daily at 18h GMT (after market close).
		{Name: "sikafinance", Spec: "0 18 * * *", Run: scheduler.Wrap(ctx, "sikafinance", marketSvc.CollectSikafinance)},
		// FluxBourse: every 30 min, 9h-15h GMT.
		{Name: "fluxbourse", Spec: "*/30 9-15 * * *", Run: scheduler.Wrap(ctx, "fluxbourse", fundamentalSvc.CollectFluxBourse)},
		// Cache cleanup: daily at 2h GMT.
		{Name: "cache_cleanup", Spec: "0 2 * * *", Run: func() { redisCache.Cleanup(ctx) }},
		// BCEAO macroeconomic data: 1st of each month at 8h GMT.
		{Name: "bceao", Spec: "0 8 1 * *", Run: scheduler.Wrap(ctx, "bceao", macroSvc.CollectBCEAO)},
		// Commodity prices (cocoa, oil): daily at 17h GMT (after markets close).
		{Name: "commodities", Spec: "0 17 * * *", Run: scheduler.Wrap(ctx, "commodities", macroSvc.CollectCommodities)},
		// Political risk (ECOWAS, GDELT): daily at 7h GMT (before market opens).
		{Name: "political_risk", Spec: "0 7 * * *", Run: scheduler.Wrap(ctx, "political_risk", macroSvc.CollectPoliticalRisk)},
	}

	cronScheduler, err := scheduler.New(jobs)
	if err != nil {
		slog.Error("failed to build scheduler", "err", err)
		os.Exit(1)
	}
	cronScheduler.Start()
	slog.Info("scheduler started, waiting for cron jobs")

	// Prometheus metrics server.
	http.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Addr: ":9090"}
	go func() {
		slog.Info("starting metrics server", "addr", ":9090")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("metrics server error", "err", err)
		}
	}()

	// Wait for interrupt signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	cronScheduler.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "err", err)
	}
	slog.Info("server exited")
}
