package internal

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Config holds all environment-sourced configuration.
type Config struct {
	DatabaseURL    string
	RedisURL       string
	GithubCSVURL   string
	SikafinanceURL string
	FluxbourseURL  string
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() Config {
	return Config{
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		RedisURL:       os.Getenv("REDIS_URL"),
		GithubCSVURL:   os.Getenv("GITHUB_CSV_URL"),
		SikafinanceURL: os.Getenv("SIKAFINANCE_URL"),
		FluxbourseURL:  os.Getenv("FLUXBOURSE_URL"),
	}
}

// Collector holds shared dependencies for all data collectors.
type Collector struct {
	DB     *pgxpool.Pool
	RDB    *redis.Client
	HTTP   *http.Client
	Config Config
}

// NewCollector creates a Collector with a sensible HTTP client.
func NewCollector(db *pgxpool.Pool, rdb *redis.Client, cfg Config) *Collector {
	return &Collector{
		DB:  db,
		RDB: rdb,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
		Config: cfg,
	}
}

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

// parseCSV reads CSV data with expected format: date,open,high,low,close,volume
// Skips the header row. Silently skips malformed rows.
func parseCSV(csvData string) ([]MarketRecord, error) {
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

	var result []MarketRecord
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

// parseMarketRow converts a single CSV row into a MarketRecord.
// Returns (record, true) on success or (zero, false) on any parse error.
func parseMarketRow(row []string) (MarketRecord, bool) {
	date, err := parseDate(row[0])
	if err != nil {
		return MarketRecord{}, false
	}

	open, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
	if err != nil {
		return MarketRecord{}, false
	}
	high, err := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
	if err != nil {
		return MarketRecord{}, false
	}
	low, err := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
	if err != nil {
		return MarketRecord{}, false
	}
	closePrice, err := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
	if err != nil {
		return MarketRecord{}, false
	}
	volume, err := strconv.ParseInt(strings.TrimSpace(row[5]), 10, 64)
	if err != nil {
		return MarketRecord{}, false
	}

	return MarketRecord{
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

// fetchURL performs an HTTP GET with exponential backoff retries and returns the response body.
func (c *Collector) fetchURL(url string) ([]byte, error) {
	var body []byte
	var finalErr error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Duration(1<<i) * time.Second) // 2s, 4s
		}
		resp, err := c.HTTP.Get(url)
		if err != nil {
			finalErr = fmt.Errorf("fetch %s: %w", url, err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			finalErr = fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
			if resp.StatusCode >= 500 { // Retry on server errors
				continue
			}
			return nil, finalErr // Fail immediately on 4xx client errors
		}

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			finalErr = fmt.Errorf("read body %s: %w", url, err)
			continue
		}

		return body, nil
	}

	return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, finalErr)
}

// batchInsertMarketData inserts OHLCV records using pgx.Batch for performance.
// Returns the number of rows inserted.
func (c *Collector) batchInsertMarketData(ctx context.Context, ticker, source string, records []MarketRecord) (int, error) {
	if len(records) == 0 {
		return 0, nil
	}

	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	batch := &pgx.Batch{}
	for _, r := range records {
		batch.Queue(
			`INSERT INTO market_data (time, ticker, open, high, low, close, volume, source)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT DO NOTHING`,
			r.Date, ticker, r.Open, r.High, r.Low, r.Close, r.Volume, source,
		)
	}

	br := tx.SendBatch(ctx, batch)
	for i := 0; i < len(records); i++ {
		if _, err := br.Exec(); err != nil {
			br.Close()
			return 0, fmt.Errorf("batch exec row %d: %w", i, err)
		}
	}
	if err := br.Close(); err != nil {
		return 0, fmt.Errorf("batch close: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return len(records), nil
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

// upsertFundamentalData upserts a single fundamental record.
// Uses INSERT ... ON CONFLICT to update existing rows.
func (c *Collector) upsertFundamentalData(ctx context.Context, rec FundamentalRecord) error {
	_, err := c.DB.Exec(ctx,
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
