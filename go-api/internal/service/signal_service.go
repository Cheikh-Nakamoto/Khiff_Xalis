package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	brvmgrpc "github.com/brvm/go-api/internal/grpc"
	"github.com/brvm/go-api/internal/model"
	"github.com/brvm/go-api/internal/repository"
	pb "github.com/brvm/go-api/proto"
	"github.com/redis/go-redis/v9"
)

// SignalService handles signal generation, caching, and Redis pub/sub.
type SignalService struct {
	MarketRepo *repository.MarketRepo
	Redis      *redis.Client
	Engine     *brvmgrpc.EngineClient // nil if engine is not available
}

const signalCacheTTL = 5 * time.Minute

// GetSignal returns a signal, checking Redis cache first, then engine, then DB.
func (s *SignalService) GetSignal(ctx context.Context, ticker string) (cached bool, result *model.SignalResult, err error) {
	// Check Redis cache
	cachedData, cacheErr := s.Redis.Get(ctx, "signal:"+ticker).Result()
	if cacheErr == nil && cachedData != "" {
		return true, &model.SignalResult{}, nil // cached flag is enough; caller uses raw string
	}

	// If engine is available, call Rust Engine via gRPC
	if s.Engine != nil {
		sr, engineErr := s.callEngineForSignal(ctx, ticker)
		if engineErr == nil {
			// Cache the result
			if data, marshalErr := json.Marshal(sr); marshalErr == nil {
				s.Redis.Set(ctx, "signal:"+ticker, data, signalCacheTTL)
			}
			return false, sr, nil
		}
		log.Printf("Engine GenerateSignal failed for %s, falling back to DB: %v", ticker, engineErr)
	}

	// Fallback: query DB
	sr, dbErr := s.MarketRepo.GetSignal(ctx, ticker)
	if dbErr != nil {
		return false, nil, fmt.Errorf("no signal data available: %w", dbErr)
	}

	// Cache the result
	if data, marshalErr := json.Marshal(sr); marshalErr == nil {
		s.Redis.Set(ctx, "signal:"+ticker, data, signalCacheTTL)
	}

	return false, sr, nil
}

// callEngineForSignal calls the Rust Engine via gRPC for a single ticker.
func (s *SignalService) callEngineForSignal(ctx context.Context, ticker string) (*model.SignalResult, error) {
	// Fetch last 90 prices, volumes, highs, lows
	prices, volumes, highs, lows, err := s.MarketRepo.GetPricesAndVolumes(ctx, ticker, 90)
	if err != nil || len(prices) == 0 {
		return nil, fmt.Errorf("no market data for %s: %w", ticker, err)
	}

	currentPrice := prices[0] // most recent

	// Fetch fundamental data (optional)
	fund, _ := s.MarketRepo.GetFundamentals(ctx, ticker)
	var pbFund *pb.FundamentalData
	if fund != nil {
		pbFund = &pb.FundamentalData{
			Per:               fund.PER,
			Roe:               fund.ROE,
			DividendYield:     fund.DividendYield,
			Eps:               fund.EPS,
			BookValuePerShare: fund.BookValuePerShare,
			DebtToEquity:      fund.DebtToEquity,
			RevenueGrowth:     fund.RevenueGrowth,
			NetProfit:         fund.NetProfit,
		}
	}

	resp, err := s.Engine.GenerateSignal(ctx, ticker, prices, volumes, highs, lows, currentPrice, pbFund)
	if err != nil {
		return nil, err
	}

	// Convert proto response to model
	sr := &model.SignalResult{
		Ticker:         model.Ticker{Symbol: resp.Ticker},
		CompositeScore: resp.CompositeScore,
		Signal:         resp.Signal,
		Confidence:     resp.Confidence,
		Reasons:        resp.Reasons,
		Timestamp:      time.Now(),
	}

	if resp.Technical != nil {
		sr.TechnicalIndicators = map[string]interface{}{
			"sma_20":          resp.Technical.Sma_20,
			"sma_50":          resp.Technical.Sma_50,
			"ema_12":          resp.Technical.Ema_12,
			"ema_26":          resp.Technical.Ema_26,
			"rsi_14":          resp.Technical.Rsi_14,
			"macd":            resp.Technical.Macd,
			"macd_signal":     resp.Technical.MacdSignal,
			"bollinger_upper": resp.Technical.BollingerUpper,
			"bollinger_lower": resp.Technical.BollingerLower,
			"atr_14":          resp.Technical.Atr_14,
			"volume_sma_20":   resp.Technical.VolumeSma_20,
		}
	}

	sr.FundamentalScores = resp.FundamentalScores

	return sr, nil
}

// GetCachedSignalString returns the raw cached signal string, or empty.
func (s *SignalService) GetCachedSignalString(ctx context.Context, ticker string) string {
	val, err := s.Redis.Get(ctx, "signal:"+ticker).Result()
	if err != nil {
		return ""
	}
	return val
}

// ScanSignals returns filtered scoring results from engine or DB fallback.
func (s *SignalService) ScanSignals(ctx context.Context, minScore float64, signalType string) ([]model.SignalResult, error) {
	// If engine is available, call Rust Engine via gRPC
	if s.Engine != nil {
		results, engineErr := s.callEngineForScan(ctx, minScore, signalType)
		if engineErr == nil {
			return results, nil
		}
		log.Printf("Engine ScanAllTickers failed, falling back to DB: %v", engineErr)
	}

	// Fallback: query DB
	return s.MarketRepo.ScanSignals(ctx, minScore, signalType)
}

// callEngineForScan calls the Rust Engine via gRPC for all BRVM tickers.
func (s *SignalService) callEngineForScan(ctx context.Context, minScore float64, signalType string) ([]model.SignalResult, error) {
	tickers := model.BRVMTickers()
	var tickerData []*pb.TickerData

	for _, t := range tickers {
		prices, volumes, highs, lows, err := s.MarketRepo.GetPricesAndVolumes(ctx, t.Symbol, 90)
		if err != nil || len(prices) == 0 {
			continue
		}

		td := &pb.TickerData{
			Symbol:       t.Symbol,
			Name:         t.Name,
			Country:      t.Country,
			Sector:       t.Sector,
			Prices:       prices,
			Volumes:      volumes,
			Highs:        highs,
			Lows:         lows,
			CurrentPrice: prices[0],
		}

		// Fetch fundamental data (optional)
		fund, _ := s.MarketRepo.GetFundamentals(ctx, t.Symbol)
		if fund != nil {
			td.Fundamental = &pb.FundamentalData{
				Per:               fund.PER,
				Roe:               fund.ROE,
				DividendYield:     fund.DividendYield,
				Eps:               fund.EPS,
				BookValuePerShare: fund.BookValuePerShare,
				DebtToEquity:      fund.DebtToEquity,
				RevenueGrowth:     fund.RevenueGrowth,
				NetProfit:         fund.NetProfit,
			}
		}

		tickerData = append(tickerData, td)
	}

	if len(tickerData) == 0 {
		return nil, fmt.Errorf("no tickers with market data")
	}

	signals, err := s.Engine.ScanAllTickers(ctx, tickerData)
	if err != nil {
		return nil, err
	}

	// Apply filters and convert to model
	var results []model.SignalResult
	for _, sig := range signals {
		if signalType != "" && sig.Signal != signalType {
			continue
		}
		if sig.CompositeScore < minScore {
			continue
		}

		sr := model.SignalResult{
			Ticker:         model.Ticker{Symbol: sig.Ticker},
			CompositeScore: sig.CompositeScore,
			Signal:         sig.Signal,
			Confidence:     sig.Confidence,
			Reasons:        sig.Reasons,
			Timestamp:      time.Now(),
		}

		if sig.Technical != nil {
			sr.TechnicalIndicators = map[string]interface{}{
				"sma_20":          sig.Technical.Sma_20,
				"sma_50":          sig.Technical.Sma_50,
				"ema_12":          sig.Technical.Ema_12,
				"ema_26":          sig.Technical.Ema_26,
				"rsi_14":          sig.Technical.Rsi_14,
				"macd":            sig.Technical.Macd,
				"macd_signal":     sig.Technical.MacdSignal,
				"bollinger_upper": sig.Technical.BollingerUpper,
				"bollinger_lower": sig.Technical.BollingerLower,
				"atr_14":          sig.Technical.Atr_14,
				"volume_sma_20":   sig.Technical.VolumeSma_20,
			}
		}

		sr.FundamentalScores = sig.FundamentalScores
		results = append(results, sr)
	}

	return results, nil
}

// SubscribeRealtime opens a Redis pub/sub channel and writes messages to the provided channel.
func (s *SignalService) SubscribeRealtime(ctx context.Context) *redis.PubSub {
	return s.Redis.Subscribe(ctx, "signals:realtime")
}
