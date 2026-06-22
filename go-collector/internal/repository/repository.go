// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

// Package repository defines the persistence interfaces used by the services.
// Concrete implementations live in sub-packages (e.g. repository/postgres).
package repository

import (
	"context"
	"time"

	"github.com/brvm/go-collector/internal/domain"
)

// MarketRepository persists OHLCV market data.
type MarketRepository interface {
	// BatchInsertMarketData inserts OHLCV records and returns the number inserted.
	BatchInsertMarketData(ctx context.Context, ticker, source string, records []domain.MarketRecord) (int, error)
}

// FundamentalRepository persists fundamental ratios.
type FundamentalRepository interface {
	UpsertFundamental(ctx context.Context, rec domain.FundamentalRecord) error
}

// MacroRepository persists macroeconomic, commodity and political-risk data.
type MacroRepository interface {
	InsertIndicators(ctx context.Context, indicators []domain.BCEAOIndicator) error
	// LatestIndicatorValue returns the most recent stored value for an indicator.
	LatestIndicatorValue(ctx context.Context, indicator string) (float64, error)
	InsertCommodityPrices(ctx context.Context, prices []domain.CommodityPrice, countries []string) error
	SeedCommodityBetas(ctx context.Context) error
	InsertPoliticalRisk(ctx context.Context, scores []domain.PoliticalRiskScore, t time.Time) error
}
