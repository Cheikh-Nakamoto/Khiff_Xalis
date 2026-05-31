package internal

import (
	"context"
	"log"
)

// collectSikafinance fetches historical OHLCV CSV data from sikafinance.com,
// parses it, and batch-inserts into market_data.
//
// Expected CSV format from sikafinance:
//
//	date,open,high,low,close,volume
//	2024-01-02,1200,1250,1180,1230,50000
//	2024-01-03,1230,1270,1210,1260,75000
//
// The SIKAFINANCE_URL env var should point to the CSV endpoint.
// Tickers are typically embedded in the URL (e.g. SIKAFINANCE_URL for each ticker)
// or the CSV contains a ticker column. We support the simplest case:
// one ticker per URL call.
func (c *Collector) CollectSikafinance(ctx context.Context) error {
	if c.Config.SikafinanceURL == "" {
		log.Println("[SIKAFINANCE] No SIKAFINANCE_URL configured, skipping")
		return nil
	}

	body, err := c.fetchURL(c.Config.SikafinanceURL)
	if err != nil {
		return err
	}

	log.Printf("[SIKAFINANCE] Downloaded %d bytes", len(body))

	// Parse the CSV — same format as GitHub (date,open,high,low,close,volume).
	records, err := parseCSV(string(body))
	if err != nil {
		return err
	}

	if len(records) == 0 {
		log.Println("[SIKAFINANCE] No valid records parsed")
		return nil
	}

	// Sikafinance typically provides data for a single ticker.
	// Derive ticker from the URL path or use a default.
	// For multi-ticker endpoints, the caller should set up multiple cron entries.
	ticker := extractTickerFromURL(c.Config.SikafinanceURL, "SNTS")

	n, err := c.batchInsertMarketData(ctx, ticker, "sikafinance", records)
	if err != nil {
		return err
	}

	log.Printf("[SIKAFINANCE] Inserted %d rows for %s", n, ticker)

	// Trigger signal re-scan
	c.RDB.Publish(ctx, "signals:scan", "sikafinance_updated")
	return nil
}
