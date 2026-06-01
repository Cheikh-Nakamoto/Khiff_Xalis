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
	"golang.org/x/sync/errgroup"
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
		var sr model.SignalResult
		if jsonErr := json.Unmarshal([]byte(cachedData), &sr); jsonErr == nil {
			return true, &sr, nil
		}
		// Cache corrupted — fall through to regenerate
		log.Printf("Cache unmarshal error for %s, regenerating: %v", ticker, cacheErr)
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

	// Fetch macroeconomic data (optional)
	macroData, _ := s.MarketRepo.GetMacroData(ctx, ticker)
	var pbMacro *pb.MacroData
	if macroData != nil {
		pbMacro = &pb.MacroData{
			Inflation:          macroData.Inflation,
			TauxDirecteur:      macroData.TauxDirecteur,
			ChangeXofEur:       macroData.ChangeXofEur,
			ChangeXofUsd:       macroData.ChangeXofUsd,
			CocoaPrice:         macroData.CocoaPrice,
			OilPrice:           macroData.OilPrice,
			CashewPrice:        macroData.CashewPrice,
			GoldPrice:          macroData.GoldPrice,
			RubberPrice:        macroData.RubberPrice,
			PalmOilPrice:       macroData.PalmOilPrice,
			PoliticalStability: macroData.PoliticalStability,
			PoliticalCrisis:    macroData.PoliticalCrisis,
			SovereignRating:    macroData.SovereignRating,
			CommodityBeta:      macroData.CommodityBeta,
		}
	}

	resp, err := s.Engine.GenerateSignal(ctx, ticker, prices, volumes, highs, lows, currentPrice, pbFund, pbMacro)
	if err != nil {
		return nil, err
	}

	// Convert proto response to model
	sr := &model.SignalResult{
		Ticker:               model.Ticker{Symbol: resp.Ticker},
		CompositeScore:       resp.CompositeScore,
		Signal:               resp.Signal,
		Confidence:           resp.Confidence,
		Reasons:              resp.Reasons,
		MacroScore:           resp.MacroScore,
		MacroReasons:         resp.MacroReasons,
		SeasonalityScore:     resp.SeasonalityScore,
		KellyFraction:        resp.KellyFraction,
		SuggestedPositionPct: resp.SuggestedPositionPct,
		Timestamp:            time.Now(),
	}

	if resp.Technical != nil {
		sr.TechnicalIndicators = map[string]interface{}{
			"sma_20":            resp.Technical.Sma_20,
			"sma_50":            resp.Technical.Sma_50,
			"ema_12":            resp.Technical.Ema_12,
			"ema_26":            resp.Technical.Ema_26,
			"rsi_14":            resp.Technical.Rsi_14,
			"macd":              resp.Technical.Macd,
			"macd_signal":     resp.Technical.MacdSignal,
			"bollinger_upper": resp.Technical.BollingerUpper,
			"bollinger_lower": resp.Technical.BollingerLower,
			"atr_14":          resp.Technical.Atr_14,
			"volume_sma_20":   resp.Technical.VolumeSma_20,
			"amihud_20":       resp.Technical.Amihud_20,
			"zero_return_ratio": resp.Technical.ZeroReturnRatio,
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
	tickerData := make([]*pb.TickerData, len(tickers))

	var eg errgroup.Group
	eg.SetLimit(10) // Limit concurrency to avoid exhausting DB connections

	for i, t := range tickers {
		i, t := i, t
		eg.Go(func() error {
			prices, volumes, highs, lows, err := s.MarketRepo.GetPricesAndVolumes(ctx, t.Symbol, 90)
			if err != nil || len(prices) == 0 {
				return nil
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

			macroData, _ := s.MarketRepo.GetMacroData(ctx, t.Symbol)
			if macroData != nil {
				td.Macro = &pb.MacroData{
					Inflation:          macroData.Inflation,
					TauxDirecteur:      macroData.TauxDirecteur,
					ChangeXofEur:       macroData.ChangeXofEur,
					ChangeXofUsd:       macroData.ChangeXofUsd,
					CocoaPrice:         macroData.CocoaPrice,
					OilPrice:           macroData.OilPrice,
					CashewPrice:        macroData.CashewPrice,
					GoldPrice:          macroData.GoldPrice,
					RubberPrice:        macroData.RubberPrice,
					PalmOilPrice:       macroData.PalmOilPrice,
					PoliticalStability: macroData.PoliticalStability,
					PoliticalCrisis:    macroData.PoliticalCrisis,
					SovereignRating:    macroData.SovereignRating,
					CommodityBeta:      macroData.CommodityBeta,
				}
			}

			tickerData[i] = td
			return nil
		})
	}

	_ = eg.Wait()

	// Filter out nil entries (tickers with errors or no data)
	var finalTickerData []*pb.TickerData
	for _, td := range tickerData {
		if td != nil {
			finalTickerData = append(finalTickerData, td)
		}
	}

	if len(finalTickerData) == 0 {
		return nil, fmt.Errorf("no tickers with market data")
	}

	signals, err := s.Engine.ScanAllTickers(ctx, finalTickerData)
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
			Ticker:               model.Ticker{Symbol: sig.Ticker},
			CompositeScore:       sig.CompositeScore,
			Signal:               sig.Signal,
			Confidence:           sig.Confidence,
			Reasons:              sig.Reasons,
			MacroScore:           sig.MacroScore,
			MacroReasons:         sig.MacroReasons,
			SeasonalityScore:     sig.SeasonalityScore,
			KellyFraction:        sig.KellyFraction,
			SuggestedPositionPct: sig.SuggestedPositionPct,
			Timestamp:            time.Now(),
		}

		if sig.Technical != nil {
			sr.TechnicalIndicators = map[string]interface{}{
				"sma_20":            sig.Technical.Sma_20,
				"sma_50":            sig.Technical.Sma_50,
				"ema_12":            sig.Technical.Ema_12,
				"ema_26":            sig.Technical.Ema_26,
				"rsi_14":            sig.Technical.Rsi_14,
				"macd":            sig.Technical.Macd,
				"macd_signal":     sig.Technical.MacdSignal,
				"bollinger_upper": sig.Technical.BollingerUpper,
				"bollinger_lower": sig.Technical.BollingerLower,
				"atr_14":          sig.Technical.Atr_14,
				"volume_sma_20":   sig.Technical.VolumeSma_20,
				"amihud_20":       sig.Technical.Amihud_20,
				"zero_return_ratio": sig.Technical.ZeroReturnRatio,
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
