// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package source

import (
	"fmt"
	"strings"
	"time"

	"github.com/brvm/go-collector/internal/domain"
	"github.com/gocolly/colly/v2"
)

// FluxBourseSource scrapes fundamental data (PER, dividend yield, ROE, EPS)
// from fluxbourse.com.
//
// FluxBourse displays a table of BRVM stocks with columns typically including:
//   - Ticker / Code
//   - PER (Price-to-Earnings Ratio)
//   - Rendement (Dividend Yield %)
//   - ROE (Return on Equity %)
//   - BPA (EPS - Earnings Per Share)
type FluxBourseSource struct {
	url string
}

// NewFluxBourseSource creates a FluxBourseSource. url is the FLUXBOURSE_URL page.
func NewFluxBourseSource(url string) *FluxBourseSource {
	return &FluxBourseSource{url: url}
}

// Configured reports whether a FluxBourse URL was provided.
func (s *FluxBourseSource) Configured() bool {
	return s.url != ""
}

// Fetch scrapes fundamental records, retrying with exponential backoff.
func (s *FluxBourseSource) Fetch() ([]domain.FundamentalRecord, error) {
	var fundamentals []domain.FundamentalRecord

	scraper := colly.NewCollector(
		colly.MaxDepth(1),
	)

	// FluxBourse typically renders data in HTML tables.
	// We look for table rows and extract columns by position.
	scraper.OnHTML("table tbody tr", func(e *colly.HTMLElement) {
		cols := e.ChildTexts("td")
		if len(cols) < 4 {
			return
		}

		// First column is typically the ticker symbol.
		ticker := strings.TrimSpace(strings.ToUpper(cols[0]))
		if ticker == "" {
			return
		}

		rec := domain.FundamentalRecord{Ticker: ticker}

		// Try to extract PER (usually 2nd column).
		if v, ok := parseFloatPtr(cols[1]); ok {
			rec.PER = v
		}

		// Try to extract dividend yield (usually 3rd column).
		if v, ok := parseFloatPtr(cols[2]); ok {
			rec.DividendYield = v
		}

		// Try to extract ROE (usually 4th column).
		if v, ok := parseFloatPtr(cols[3]); ok {
			rec.ROE = v
		}

		// EPS if a 5th column is present.
		if len(cols) >= 5 {
			if v, ok := parseFloatPtr(cols[4]); ok {
				rec.EPS = v
			}
		}

		fundamentals = append(fundamentals, rec)
	})

	// Also try div-based layouts (some sites use divs instead of tables).
	scraper.OnHTML(".stock-row, .cote-row, [data-ticker]", func(e *colly.HTMLElement) {
		ticker := strings.TrimSpace(strings.ToUpper(
			e.Attr("data-ticker"),
		))
		if ticker == "" {
			ticker = strings.TrimSpace(strings.ToUpper(e.ChildText(".ticker, .symbol, .code")))
		}
		if ticker == "" {
			return
		}

		rec := domain.FundamentalRecord{Ticker: ticker}

		if v, ok := parseFloatFromElement(e, ".per, .pe-ratio"); ok {
			rec.PER = v
		}
		if v, ok := parseFloatFromElement(e, ".yield, .dividend-yield, .rendement"); ok {
			rec.DividendYield = v
		}
		if v, ok := parseFloatFromElement(e, ".roe"); ok {
			rec.ROE = v
		}
		if v, ok := parseFloatFromElement(e, ".eps, .bpa"); ok {
			rec.EPS = v
		}

		fundamentals = append(fundamentals, rec)
	})

	var err error
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(1<<i) * time.Second)
		}
		err = scraper.Visit(s.url)
		if err == nil {
			return fundamentals, nil
		}
	}

	return nil, fmt.Errorf("fluxbourse failed after %d retries: %w", maxRetries, err)
}
