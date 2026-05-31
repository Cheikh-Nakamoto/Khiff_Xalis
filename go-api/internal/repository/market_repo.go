package repository

import (
	"context"
	"time"

	"github.com/brvm/go-api/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MarketRepo handles market-data queries.
type MarketRepo struct {
	DB *pgxpool.Pool
}

// GetMarketData returns OHLCV bars for a ticker over the past N days.
func (r *MarketRepo) GetMarketData(ctx context.Context, ticker string, days int) ([]model.MarketDataPoint, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT time, open, high, low, close, volume, source
		 FROM market_data
		 WHERE ticker = $1 AND time > NOW() - make_interval(days => $2)
		 ORDER BY time DESC`,
		ticker, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []model.MarketDataPoint
	for rows.Next() {
		var d model.MarketDataPoint
		var t time.Time
		var source string
		if err := rows.Scan(&t, &d.Open, &d.High, &d.Low, &d.Close, &d.Volume, &source); err != nil {
			continue
		}
		d.Date = t.Format("2006-01-02")
		d.Source = source
		if len(data) > 0 && data[len(data)-1].Close != 0 {
			d.DailyReturn = (d.Close - data[len(data)-1].Close) / data[len(data)-1].Close * 100
		}
		data = append(data, d)
	}
	return data, nil
}

// GetLatestData returns the most recent bar for a ticker.
func (r *MarketRepo) GetLatestData(ctx context.Context, ticker string) (*model.MarketDataPoint, error) {
	var d model.MarketDataPoint
	var t time.Time
	var source string
	err := r.DB.QueryRow(ctx,
		`SELECT time, open, high, low, close, volume, source
		 FROM market_data WHERE ticker = $1
		 ORDER BY time DESC LIMIT 1`, ticker).Scan(
		&t, &d.Open, &d.High, &d.Low, &d.Close, &d.Volume, &source)
	if err != nil {
		return nil, err
	}
	d.Date = t.Format("2006-01-02")
	d.Source = source
	return &d, nil
}

// GetFundamentals returns fundamental data for a ticker.
func (r *MarketRepo) GetFundamentals(ctx context.Context, ticker string) (*model.FundamentalData, error) {
	var f model.FundamentalData
	err := r.DB.QueryRow(ctx,
		`SELECT ticker, per, roe, dividend_yield, eps, book_value_per_share,
		        debt_to_equity, revenue_growth, net_profit, updated_at
		 FROM fundamental_data WHERE ticker = $1`, ticker).Scan(
		&f.Ticker, &f.PER, &f.ROE, &f.DividendYield, &f.EPS,
		&f.BookValuePerShare, &f.DebtToEquity, &f.RevenueGrowth, &f.NetProfit, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// GetSignal returns the latest scoring result for a ticker.
func (r *MarketRepo) GetSignal(ctx context.Context, ticker string) (*model.SignalResult, error) {
	var sr model.SignalResult
	var tickerSymbol string
	var reasons []string
	err := r.DB.QueryRow(ctx,
		`SELECT ticker, composite_score, signal_type, confidence, reasons, created_at
		 FROM scoring_results WHERE ticker = $1
		 ORDER BY created_at DESC LIMIT 1`, ticker).Scan(
		&tickerSymbol, &sr.CompositeScore, &sr.Signal, &sr.Confidence, &reasons, &sr.Timestamp)
	if err != nil {
		return nil, err
	}
	sr.Ticker = model.Ticker{Symbol: tickerSymbol}
	sr.Reasons = reasons
	return &sr, nil
}

// ScanSignals returns scoring results filtered by min score and optional signal type.
func (r *MarketRepo) ScanSignals(ctx context.Context, minScore float64, signalType string) ([]model.SignalResult, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT ticker, composite_score, signal_type, confidence, reasons, created_at
		 FROM scoring_results
		 WHERE composite_score >= $1
		   AND ($2 = '' OR signal_type = $2)
		 ORDER BY composite_score DESC`,
		minScore, signalType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.SignalResult
	for rows.Next() {
		var sr model.SignalResult
		var tickerSymbol string
		var reasons []string
		if err := rows.Scan(&tickerSymbol, &sr.CompositeScore, &sr.Signal, &sr.Confidence, &reasons, &sr.Timestamp); err != nil {
			continue
		}
		sr.Ticker = model.Ticker{Symbol: tickerSymbol}
		sr.Reasons = reasons
		results = append(results, sr)
	}
	return results, nil
}

// CurrentPrice returns the latest close price for a ticker (used for portfolio valuation).
func (r *MarketRepo) CurrentPrice(ctx context.Context, ticker string) (float64, error) {
	var price float64
	err := r.DB.QueryRow(ctx,
		"SELECT close FROM market_data WHERE ticker = $1 ORDER BY time DESC LIMIT 1",
		ticker).Scan(&price)
	return price, err
}

// GetPricesAndVolumes returns the last N close prices, volumes, highs, and lows for a ticker.
// Used by the gRPC engine client for signal generation.
func (r *MarketRepo) GetPricesAndVolumes(ctx context.Context, ticker string, limit int) (prices []float64, volumes []int64, highs []float64, lows []float64, err error) {
	rows, err := r.DB.Query(ctx,
		`SELECT close, volume, high, low FROM market_data
		 WHERE ticker = $1 ORDER BY time DESC LIMIT $2`, ticker, limit)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p float64
		var v int64
		var h float64
		var l float64
		if err := rows.Scan(&p, &v, &h, &l); err != nil {
			continue
		}
		prices = append(prices, p)
		volumes = append(volumes, v)
		highs = append(highs, h)
		lows = append(lows, l)
	}
	return prices, volumes, highs, lows, nil
}
