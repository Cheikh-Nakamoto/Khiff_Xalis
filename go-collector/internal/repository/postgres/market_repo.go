// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

// Package postgres provides PostgreSQL-backed implementations of the
// repository interfaces.
package postgres

import (
	"context"
	"fmt"

	"github.com/brvm/go-collector/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MarketRepo persists OHLCV market data into the market_data hypertable.
type MarketRepo struct {
	db *pgxpool.Pool
}

// NewMarketRepo creates a MarketRepo.
func NewMarketRepo(db *pgxpool.Pool) *MarketRepo {
	return &MarketRepo{db: db}
}

// BatchInsertMarketData inserts OHLCV records using pgx.Batch for performance.
// Returns the number of rows inserted.
func (r *MarketRepo) BatchInsertMarketData(ctx context.Context, ticker, source string, records []domain.MarketRecord) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	batch := &pgx.Batch{}
	for _, rec := range records {
		batch.Queue(
			`INSERT INTO market_data (time, ticker, open, high, low, close, volume, source)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT DO NOTHING`,
			rec.Date, ticker, rec.Open, rec.High, rec.Low, rec.Close, rec.Volume, source,
		)
	}

	br := tx.SendBatch(ctx, batch)
	for i := 0; i < len(records); i++ {
		if _, err := br.Exec(); err != nil {
			br.Close()
			return 0, fmt.Errorf("batch exec row %d: %w", i, err)
		}
	}
	if err := br.Close(); err != nil {
		return 0, fmt.Errorf("batch close: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return len(records), nil
}
