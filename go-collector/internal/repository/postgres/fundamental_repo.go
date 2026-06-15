// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package postgres

import (
	"context"

	"github.com/brvm/go-collector/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FundamentalRepo persists fundamental ratios into the fundamental_data table.
type FundamentalRepo struct {
	db *pgxpool.Pool
}

// NewFundamentalRepo creates a FundamentalRepo.
func NewFundamentalRepo(db *pgxpool.Pool) *FundamentalRepo {
	return &FundamentalRepo{db: db}
}

// UpsertFundamental upserts a single fundamental record.
// Uses INSERT ... ON CONFLICT to update existing rows.
func (r *FundamentalRepo) UpsertFundamental(ctx context.Context, rec domain.FundamentalRecord) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO fundamental_data (ticker, per, roe, dividend_yield, eps, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())
		 ON CONFLICT (ticker) DO UPDATE SET
		   per = EXCLUDED.per,
		   roe = EXCLUDED.roe,
		   dividend_yield = EXCLUDED.dividend_yield,
		   eps = EXCLUDED.eps,
		   updated_at = NOW()`,
		rec.Ticker, rec.PER, rec.ROE, rec.DividendYield, rec.EPS,
	)
	return err
}
