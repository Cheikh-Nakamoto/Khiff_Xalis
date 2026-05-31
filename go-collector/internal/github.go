package internal

import (
	"context"
	"fmt"
	"log"
)

// BRVM tickers available in the brvm-data-public GitHub repo.
var brvmTickers = []string{
	"SNTS", "ORAC", "SGBC", "ECOC", "PALC", "SPHC", "SMB", "SOGC",
	"SLBC", "TTLC", "TTLS", "UNLC", "FILC", "SAFC", "SICC", "SDCC", "SDSC",
	"SCRC", "SVC", "STAC", "UNXC", "SHEC", "ONTBF", "ORGT", "NEST", "NSIA",
}

// CollectGithubCSV fetches OHLCV CSV data from the brvm-data-public GitHub repo
// and batch-inserts into market_data.
func (c *Collector) CollectGithubCSV(ctx context.Context) error {
	var totalInserted int

	for _, ticker := range brvmTickers {
		url := fmt.Sprintf("%s/%s.csv", c.Config.GithubCSVURL, ticker)

		body, err := c.fetchURL(url)
		if err != nil {
			log.Printf("[GITHUB] Fetch %s: %v", ticker, err)
			continue
		}

		records, err := parseCSV(string(body))
		if err != nil {
			log.Printf("[GITHUB] Parse %s: %v", ticker, err)
			continue
		}

		n, err := c.batchInsertMarketData(ctx, ticker, "github_brvm_data_public", records)
		if err != nil {
			log.Printf("[GITHUB] Insert %s: %v", ticker, err)
			continue
		}
		totalInserted += n
	}

	log.Printf("[GITHUB] Collection complete: %d rows inserted across %d tickers", totalInserted, len(brvmTickers))

	// Publish scan trigger so the engine knows new data arrived.
	c.RDB.Publish(ctx, "signals:scan", "github_updated")
	return nil
}
