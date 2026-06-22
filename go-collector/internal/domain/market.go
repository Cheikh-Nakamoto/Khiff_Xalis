// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

// Package domain holds the pure data types shared across the collector.
// It must not import any infrastructure (database, redis, http, scraping).
package domain

import "time"

// MarketRecord holds one parsed OHLCV row.
type MarketRecord struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// FundamentalRecord holds scraped fundamental data for one ticker.
type FundamentalRecord struct {
	Ticker        string
	PER           *float64
	ROE           *float64
	DividendYield *float64
	EPS           *float64
}
