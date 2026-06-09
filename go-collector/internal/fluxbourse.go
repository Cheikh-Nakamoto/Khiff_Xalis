package internal

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

// collectFluxBourse scrapes fundamental data (PER, dividend yield, ROE, EPS)
// from fluxbourse.com and upserts into the fundamental_data table.
//
// FluxBourse displays a table of BRVM stocks with columns typically including:
//   - Ticker / Code
//   - PER (Price-to-Earnings Ratio)
//   - Rendement (Dividend Yield %)
//   - ROE (Return on Equity %)
//   - BPA (EPS - Earnings Per Share)
//
// The FLUXBOURSE_URL env var should point to the page containing the table.
func (c *Collector) CollectFluxBourse(ctx context.Context) error {
	if c.Config.FluxbourseURL == "" {
		log.Println("[FLUXBOURSE] No FLUXBOURSE_URL configured, skipping")
		return nil
	}

	var fundamentals []FundamentalRecord

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

		rec := FundamentalRecord{Ticker: ticker}

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

		rec := FundamentalRecord{Ticker: ticker}

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
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(1<<i) * time.Second)
		}
		err = scraper.Visit(c.Config.FluxbourseURL)
		if err == nil {
			break
		}
		log.Printf("[FLUXBOURSE] Scrape attempt %d failed: %v", i+1, err)
	}

	if err != nil {
		return fmt.Errorf("fluxbourse failed after %d retries: %w", maxRetries, err)
	}

	log.Printf("[FLUXBOURSE] Scraped %d fundamental records", len(fundamentals))

	// Upsert each record into fundamental_data.
	var inserted int
	for _, rec := range fundamentals {
		if err := c.upsertFundamentalData(ctx, rec); err != nil {
			log.Printf("[FLUXBOURSE] Upsert %s: %v", rec.Ticker, err)
			continue
		}
		inserted++
	}

	log.Printf("[FLUXBOURSE] Upserted %d/%d records", inserted, len(fundamentals))

	// Publish update notification.
	c.RDB.Publish(ctx, "signals:scan", "fluxbourse_updated")
	return nil
}

// parseFloatPtr parses a string into a *float64.
// Returns (nil, false) if the value is empty, dash, or not a valid number.
func parseFloatPtr(s string) (*float64, bool) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "%", "")

	if s == "" || s == "-" || s == "N/A" || s == "n/a" {
		return nil, false
	}

	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, false
	}
	return &v, true
}

// parseFloatFromElement extracts text from a CSS selector and parses it.
func parseFloatFromElement(e *colly.HTMLElement, selector string) (*float64, bool) {
	text := e.ChildText(selector)
	if text == "" {
		return nil, false
	}
	return parseFloatPtr(text)
}
