// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package source

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/brvm/go-collector/internal/domain"
	"github.com/gocolly/colly/v2"
)

// parseCSV reads CSV data with expected format: date,open,high,low,close,volume
// Skips the header row. Silently skips malformed rows.
func parseCSV(csvData string) ([]domain.MarketRecord, error) {
	reader := csv.NewReader(strings.NewReader(csvData))
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // allow variable-length rows (malformed ones get skipped)

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv parse error: %w", err)
	}

	if len(records) == 0 {
		return nil, nil
	}

	var result []domain.MarketRecord
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 6 {
			continue // skip rows with insufficient columns
		}

		rec, ok := parseMarketRow(row)
		if !ok {
			continue
		}
		result = append(result, rec)
	}

	return result, nil
}

// parseMarketRow converts a single CSV row into a domain.MarketRecord.
// Returns (record, true) on success or (zero, false) on any parse error.
func parseMarketRow(row []string) (domain.MarketRecord, bool) {
	date, err := parseDate(row[0])
	if err != nil {
		return domain.MarketRecord{}, false
	}

	open, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
	if err != nil {
		return domain.MarketRecord{}, false
	}
	high, err := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
	if err != nil {
		return domain.MarketRecord{}, false
	}
	low, err := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
	if err != nil {
		return domain.MarketRecord{}, false
	}
	closePrice, err := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
	if err != nil {
		return domain.MarketRecord{}, false
	}
	volume, err := strconv.ParseInt(strings.TrimSpace(row[5]), 10, 64)
	if err != nil {
		return domain.MarketRecord{}, false
	}

	return domain.MarketRecord{
		Date:   date,
		Open:   open,
		High:   high,
		Low:    low,
		Close:  closePrice,
		Volume: volume,
	}, true
}

// parseDate tries common date formats used in BRVM CSV files.
func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
		"2006/01/02",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date: %q", s)
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

// extractTickerFromURL attempts to find a BRVM ticker symbol in the URL path.
// Falls back to defaultTicker if none is found.
func extractTickerFromURL(url, defaultTicker string) string {
	upper := strings.ToUpper(url)
	for _, t := range brvmTickers {
		if strings.Contains(upper, t) {
			return t
		}
	}
	return defaultTicker
}
