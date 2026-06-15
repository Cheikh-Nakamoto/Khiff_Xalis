// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package source

import (
	"context"

	"github.com/brvm/go-collector/internal/domain"
)

// SikafinanceSource fetches historical OHLCV CSV data from sikafinance.com.
//
// Expected CSV format (same as GitHub): date,open,high,low,close,volume.
// Sikafinance typically provides data for a single ticker per URL; tickers are
// embedded in the URL. For multi-ticker endpoints set up multiple cron entries.
type SikafinanceSource struct {
	url  string
	http *HTTPClient
}

// NewSikafinanceSource creates a SikafinanceSource. url is the SIKAFINANCE_URL endpoint.
func NewSikafinanceSource(url string, http *HTTPClient) *SikafinanceSource {
	return &SikafinanceSource{url: url, http: http}
}

// Configured reports whether a Sikafinance URL was provided.
func (s *SikafinanceSource) Configured() bool {
	return s.url != ""
}

// Ticker derives the ticker for the configured URL, defaulting to "SNTS".
func (s *SikafinanceSource) Ticker() string {
	return extractTickerFromURL(s.url, "SNTS")
}

// Fetch downloads and parses the OHLCV CSV from the configured Sikafinance URL.
func (s *SikafinanceSource) Fetch(ctx context.Context) ([]domain.MarketRecord, error) {
	body, err := s.http.Fetch(ctx, s.url)
	if err != nil {
		return nil, err
	}
	return parseCSV(string(body))
}
