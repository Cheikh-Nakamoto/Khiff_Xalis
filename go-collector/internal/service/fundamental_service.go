// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package service

import (
	"context"
	"log/slog"

	"github.com/brvm/go-collector/internal/repository"
	"github.com/brvm/go-collector/internal/source"
)

// FundamentalService collects fundamental ratios from FluxBourse.
type FundamentalService struct {
	flux *source.FluxBourseSource
	repo repository.FundamentalRepository
	pub  Publisher
}

// NewFundamentalService creates a FundamentalService.
func NewFundamentalService(flux *source.FluxBourseSource, repo repository.FundamentalRepository, pub Publisher) *FundamentalService {
	return &FundamentalService{flux: flux, repo: repo, pub: pub}
}

// CollectFluxBourse scrapes fundamental data and upserts it into fundamental_data.
func (s *FundamentalService) CollectFluxBourse(ctx context.Context) error {
	if !s.flux.Configured() {
		slog.Info("fluxbourse skipped: no FLUXBOURSE_URL configured")
		return nil
	}

	fundamentals, err := s.flux.Fetch()
	if err != nil {
		return err
	}

	slog.Info("fluxbourse scraped", "records", len(fundamentals))

	var inserted int
	for _, rec := range fundamentals {
		if err := s.repo.UpsertFundamental(ctx, rec); err != nil {
			slog.Error("fluxbourse upsert failed", "ticker", rec.Ticker, "err", err)
			continue
		}
		inserted++
	}

	slog.Info("fluxbourse collection complete", "upserted", inserted, "scraped", len(fundamentals))

	if err := s.pub.Publish(ctx, signalsScanChannel, "fluxbourse_updated"); err != nil {
		slog.Error("fluxbourse publish failed", "err", err)
	}
	return nil
}
