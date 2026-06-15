// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package source

import (
	"context"
	"fmt"

	"github.com/brvm/go-collector/internal/domain"
)

// brvmTickers lists the BRVM tickers available in the brvm-data-public GitHub repo.
var brvmTickers = []string{
	"SNTS", "ORAC", "SGBC", "ECOC", "PALC", "SPHC", "SMB", "SOGC",
	"SLBC", "TTLC", "TTLS", "UNLC", "FILC", "SAFC", "SICC", "SDCC", "SDSC",
	"SCRC", "SVC", "STAC", "UNXC", "SHEC", "ONTBF", "ORGT", "NEST", "NSIA",
}

// GithubSource fetches OHLCV CSV data from the brvm-data-public GitHub repo.
type GithubSource struct {
	baseURL string
	http    *HTTPClient
}

// NewGithubSource creates a GithubSource. baseURL is the GITHUB_CSV_URL prefix.
func NewGithubSource(baseURL string, http *HTTPClient) *GithubSource {
	return &GithubSource{baseURL: baseURL, http: http}
}

// Tickers returns the list of tickers this source knows how to fetch.
func (s *GithubSource) Tickers() []string {
	return brvmTickers
}

// Fetch downloads and parses the OHLCV CSV for a single ticker.
func (s *GithubSource) Fetch(ctx context.Context, ticker string) ([]domain.MarketRecord, error) {
	url := fmt.Sprintf("%s/%s.csv", s.baseURL, ticker)

	body, err := s.http.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", ticker, err)
	}

	records, err := parseCSV(string(body))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", ticker, err)
	}

	return records, nil
}
