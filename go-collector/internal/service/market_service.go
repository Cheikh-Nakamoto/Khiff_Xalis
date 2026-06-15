// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package service

import (
	"context"
	"log/slog"

	"github.com/brvm/go-collector/internal/domain"
	"github.com/brvm/go-collector/internal/repository"
)

// GithubFetcher fetches OHLCV records per ticker (implemented by *source.GithubSource).
type GithubFetcher interface {
	Tickers() []string
	Fetch(ctx context.Context, ticker string) ([]domain.MarketRecord, error)
}

// SikaFetcher fetches OHLCV records from a configured URL (implemented by *source.SikafinanceSource).
type SikaFetcher interface {
	Configured() bool
	Ticker() string
	Fetch(ctx context.Context) ([]domain.MarketRecord, error)
}

// MarketService collects OHLCV market data from GitHub and Sikafinance.
type MarketService struct {
	github GithubFetcher
	sika   SikaFetcher
	repo   repository.MarketRepository
	pub    Publisher
}

// NewMarketService creates a MarketService.
func NewMarketService(github GithubFetcher, sika SikaFetcher, repo repository.MarketRepository, pub Publisher) *MarketService {
	return &MarketService{github: github, sika: sika, repo: repo, pub: pub}
}

// CollectGithubCSV fetches OHLCV CSV data for every known ticker and batch-inserts it.
func (s *MarketService) CollectGithubCSV(ctx context.Context) error {
	tickers := s.github.Tickers()
	var totalInserted int

	for _, ticker := range tickers {
		records, err := s.github.Fetch(ctx, ticker)
		if err != nil {
			slog.Error("github fetch failed", "ticker", ticker, "err", err)
			continue
		}

		n, err := s.repo.BatchInsertMarketData(ctx, ticker, "github_brvm_data_public", records)
		if err != nil {
			slog.Error("github insert failed", "ticker", ticker, "err", err)
			continue
		}
		totalInserted += n
	}

	slog.Info("github collection complete", "rows", totalInserted, "tickers", len(tickers))

	// Publish scan trigger so the engine knows new data arrived.
	if err := s.pub.Publish(ctx, signalsScanChannel, "github_updated"); err != nil {
		slog.Error("github publish failed", "err", err)
	}
	return nil
}

// CollectSikafinance fetches OHLCV CSV from the configured Sikafinance URL and inserts it.
func (s *MarketService) CollectSikafinance(ctx context.Context) error {
	if !s.sika.Configured() {
		slog.Info("sikafinance skipped: no SIKAFINANCE_URL configured")
		return nil
	}

	records, err := s.sika.Fetch(ctx)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		slog.Info("sikafinance: no valid records parsed")
		return nil
	}

	ticker := s.sika.Ticker()
	n, err := s.repo.BatchInsertMarketData(ctx, ticker, "sikafinance", records)
	if err != nil {
		return err
	}

	slog.Info("sikafinance collection complete", "rows", n, "ticker", ticker)

	if err := s.pub.Publish(ctx, signalsScanChannel, "sikafinance_updated"); err != nil {
		slog.Error("sikafinance publish failed", "err", err)
	}
	return nil
}
