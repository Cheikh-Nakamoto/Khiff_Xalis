// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/brvm/go-collector/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MacroRepo persists macroeconomic, commodity and political-risk data into
// the macroeconomic_data table (and ticker_commodity_map for betas).
type MacroRepo struct {
	db *pgxpool.Pool
}

// NewMacroRepo creates a MacroRepo.
func NewMacroRepo(db *pgxpool.Pool) *MacroRepo {
	return &MacroRepo{db: db}
}

// InsertIndicators inserts macroeconomic indicators into the database.
func (r *MacroRepo) InsertIndicators(ctx context.Context, indicators []domain.BCEAOIndicator) error {
	batch := &pgx.Batch{}

	for _, ind := range indicators {
		batch.Queue(
			`INSERT INTO macroeconomic_data (time, country, indicator, value, unit, source, confidence)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			ind.Date, ind.Country, ind.Indicator, ind.Value, ind.Unit, "BCEAO/WorldBank", 0.9,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert row %d: %w", i, err)
		}
	}

	return nil
}

// LatestIndicatorValue returns the most recent stored value for an indicator.
func (r *MacroRepo) LatestIndicatorValue(ctx context.Context, indicator string) (float64, error) {
	var value float64
	err := r.db.QueryRow(ctx,
		`SELECT value FROM macroeconomic_data
		 WHERE indicator = $1
		 ORDER BY time DESC LIMIT 1`, indicator).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// InsertCommodityPrices inserts commodity prices for all countries.
func (r *MacroRepo) InsertCommodityPrices(ctx context.Context, prices []domain.CommodityPrice, countries []string) error {
	batch := &pgx.Batch{}

	for _, price := range prices {
		for _, country := range countries {
			batch.Queue(
				`INSERT INTO macroeconomic_data (time, country, indicator, value, unit, source, confidence)
				 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				price.Date, country, price.Commodity, price.Price, price.Unit, price.Source, 0.95,
			)
		}
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert row %d: %w", i, err)
		}
	}

	return nil
}

// SeedCommodityBetas populates the ticker_commodity_map table with predefined
// sensitivities (e.g. PALC rubber/palm oil exposure, SMB oil exposure).
func (r *MacroRepo) SeedCommodityBetas(ctx context.Context) error {
	betas := []struct {
		Ticker    string
		Commodity string
		Beta      float64
	}{
		{"PALC", "palm_oil_price", 1.25},
		{"PALC", "rubber_price", 0.35},
		{"SPHC", "rubber_price", 1.45},
		{"SOGC", "rubber_price", 1.15},
		{"SOGC", "palm_oil_price", 0.55},
		{"SMB", "oil_price", 1.30},
		{"NEST", "cocoa_price", -0.45}, // Consumer of cocoa, high prices are a drag
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, b := range betas {
		_, err := tx.Exec(ctx,
			`INSERT INTO ticker_commodity_map (ticker, commodity, beta, updated_at)
			 VALUES ($1, $2, $3, NOW())
			 ON CONFLICT (ticker, commodity) DO UPDATE SET
			   beta = EXCLUDED.beta,
			   updated_at = NOW()`,
			b.Ticker, b.Commodity, b.Beta,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// InsertPoliticalRisk inserts evaluated risk scores (stability + crisis flag).
func (r *MacroRepo) InsertPoliticalRisk(ctx context.Context, scores []domain.PoliticalRiskScore, t time.Time) error {
	batch := &pgx.Batch{}

	for _, s := range scores {
		// 1. Insert political_stability
		batch.Queue(
			`INSERT INTO macroeconomic_data (time, country, indicator, value, unit, source, confidence)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			t, s.Country, "political_stability", s.PoliticalStability, "Score", s.Source, 0.85,
		)

		// 2. Insert political_crisis (1.0 = Active Crisis, 0.0 = Stable)
		var crisisVal float64
		if s.PoliticalCrisis {
			crisisVal = 1.0
		}
		batch.Queue(
			`INSERT INTO macroeconomic_data (time, country, indicator, value, unit, source, confidence)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			t, s.Country, "political_crisis", crisisVal, "Status", s.Source, 0.90,
		)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert row %d: %w", i, err)
		}
	}

	return nil
}
