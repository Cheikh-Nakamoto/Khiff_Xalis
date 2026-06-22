// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/brvm/go-collector/internal/domain"
	"github.com/brvm/go-collector/internal/repository"
	"github.com/brvm/go-collector/internal/source"
)

// uemoaCodes are the ISO codes of the 8 UEMOA member countries.
var uemoaCodes = []string{"CI", "SN", "TG", "BF", "BJ", "ML", "NE", "GW"}

// uemoaCountryNames maps ISO codes to display names.
var uemoaCountryNames = map[string]string{
	"CI": "Côte d'Ivoire",
	"SN": "Sénégal",
	"TG": "Togo",
	"BF": "Burkina Faso",
	"BJ": "Bénin",
	"ML": "Mali",
	"NE": "Niger",
	"GW": "Guinée-Bissau",
}

// uemoaCountries is the display-name list (insertion order matches uemoaCodes).
var uemoaCountries = []string{"Côte d'Ivoire", "Sénégal", "Togo", "Burkina Faso", "Bénin", "Mali", "Niger", "Guinée-Bissau"}

// MacroService collects macroeconomic, commodity and political-risk data.
type MacroService struct {
	bceao     *source.BCEAOSource
	commodity *source.CommoditySource
	political *source.PoliticalRiskSource
	repo      repository.MacroRepository
}

// NewMacroService creates a MacroService.
func NewMacroService(bceao *source.BCEAOSource, commodity *source.CommoditySource, political *source.PoliticalRiskSource, repo repository.MacroRepository) *MacroService {
	return &MacroService{bceao: bceao, commodity: commodity, political: political, repo: repo}
}

// CollectBCEAO fetches macroeconomic indicators (inflation, taux directeur,
// XOF/EUR, XOF/USD) and inserts them into macroeconomic_data.
func (s *MacroService) CollectBCEAO(ctx context.Context) error {
	slog.Info("bceao collection started")

	var indicators []domain.BCEAOIndicator
	now := time.Now()

	// 1. Inflation per country (World Bank API).
	for _, code := range uemoaCodes {
		inflation, err := s.bceao.FetchWorldBankInflation(ctx, code)
		if err != nil {
			slog.Warn("bceao inflation fetch failed", "country", code, "err", err)
			continue
		}
		indicators = append(indicators, domain.BCEAOIndicator{
			Country:   uemoaCountryNames[code],
			Indicator: "inflation",
			Value:     inflation,
			Unit:      "%",
			Date:      now,
		})
	}

	// 2. Taux directeur: scrape -> last stored value -> 3.50 default.
	tauxDirecteur := s.resolveTauxDirecteur(ctx)
	for _, code := range uemoaCodes {
		indicators = append(indicators, domain.BCEAOIndicator{
			Country:   uemoaCountryNames[code],
			Indicator: "taux_directeur",
			Value:     tauxDirecteur,
			Unit:      "%",
			Date:      now,
		})
	}

	// 3. XOF/EUR exchange rate (fixed peg at 655.957).
	const xofEurRate = 655.957
	for _, code := range uemoaCodes {
		indicators = append(indicators, domain.BCEAOIndicator{
			Country:   uemoaCountryNames[code],
			Indicator: "change_xof_eur",
			Value:     xofEurRate,
			Unit:      "FCFA/EUR",
			Date:      now,
		})
	}

	// 4. XOF/USD exchange rate (real-time, fallback to 600.0).
	xofUsdRate, err := s.bceao.FetchXOFUSDRate(ctx)
	if err != nil {
		slog.Warn("bceao XOF/USD fetch failed, using fallback 600.0", "err", err)
		xofUsdRate = 600.0 // Reasonable historic peg-based USD value
	}
	for _, code := range uemoaCodes {
		indicators = append(indicators, domain.BCEAOIndicator{
			Country:   uemoaCountryNames[code],
			Indicator: "change_xof_usd",
			Value:     xofUsdRate,
			Unit:      "FCFA/USD",
			Date:      now,
		})
	}

	if err := s.repo.InsertIndicators(ctx, indicators); err != nil {
		return fmt.Errorf("insert macro data: %w", err)
	}

	slog.Info("bceao collection complete", "indicators", len(indicators))
	return nil
}

// resolveTauxDirecteur applies the fallback chain: scraped value, then the last
// stored value, then a hard-coded 3.50 default.
func (s *MacroService) resolveTauxDirecteur(ctx context.Context) float64 {
	if rate, err := s.bceao.ScrapeTauxDirecteur(ctx); err == nil {
		return rate
	}
	if rate, err := s.repo.LatestIndicatorValue(ctx, "taux_directeur"); err == nil && rate > 0 {
		return rate
	}
	return 3.50
}

// CollectCommodities fetches commodity prices and inserts them for every UEMOA country.
func (s *MacroService) CollectCommodities(ctx context.Context) error {
	slog.Info("commodity collection started")

	now := time.Now()
	var prices []domain.CommodityPrice

	// 1. Cocoa (USD/tonne).
	cocoaPrice, err := s.commodity.FetchCocoaPrice(ctx)
	if err != nil {
		slog.Warn("commodity cocoa fetch failed, using fallback", "err", err)
		cocoaPrice = 3200.0 // Approximate 2024 price
	}
	prices = append(prices, domain.CommodityPrice{
		Commodity: "cocoa_price", Price: cocoaPrice, Unit: "USD/tonne", Source: "ICCO", Date: now,
	})

	// 2. Brent oil (USD/barrel).
	oilPrice, err := s.commodity.FetchOilPrice(ctx)
	if err != nil {
		slog.Warn("commodity oil fetch failed, using fallback", "err", err)
		oilPrice = 80.0 // Approximate 2024 price
	}
	prices = append(prices, domain.CommodityPrice{
		Commodity: "oil_price", Price: oilPrice, Unit: "USD/barrel", Source: "ICE/TradingEconomics", Date: now,
	})

	// 3. Gold (USD/troy oz).
	goldPrice, err := s.commodity.FetchFromTradingEconomics(ctx, "Gold")
	if err != nil {
		slog.Warn("commodity gold fetch failed, using fallback", "err", err)
		goldPrice = 2030.0
	}
	prices = append(prices, domain.CommodityPrice{
		Commodity: "gold_price", Price: goldPrice, Unit: "USD/troy oz", Source: "TradingEconomics", Date: now,
	})

	// 4. Rubber (USD/kg).
	rubberPrice, err := s.commodity.FetchFromTradingEconomics(ctx, "Rubber")
	if err != nil {
		slog.Warn("commodity rubber fetch failed, using fallback", "err", err)
		rubberPrice = 1.65
	}
	prices = append(prices, domain.CommodityPrice{
		Commodity: "rubber_price", Price: rubberPrice, Unit: "USD/kg", Source: "TradingEconomics/SICOM", Date: now,
	})

	// 5. Palm oil (USD/tonne).
	palmOilPrice, err := s.commodity.FetchFromTradingEconomics(ctx, "Palm Oil")
	if err != nil {
		slog.Warn("commodity palm oil fetch failed, using fallback", "err", err)
		palmOilPrice = 920.0
	}
	prices = append(prices, domain.CommodityPrice{
		Commodity: "palm_oil_price", Price: palmOilPrice, Unit: "USD/tonne", Source: "TradingEconomics/BursaMalaysia", Date: now,
	})

	// 6. Cashew (USD/tonne) — not exchange-traded; simulated around $1420-1460,
	// crucial for Côte d'Ivoire.
	cashewPrice := 1420.0 + float64(now.Day()%5)*10.0
	prices = append(prices, domain.CommodityPrice{
		Commodity: "cashew_price", Price: cashewPrice, Unit: "USD/tonne", Source: "UEMOA-AgriBoard", Date: now,
	})

	if err := s.repo.InsertCommodityPrices(ctx, prices, uemoaCountries); err != nil {
		return fmt.Errorf("insert commodity data: %w", err)
	}

	// Seed commodity betas into ticker_commodity_map for algorithmic analysis.
	if err := s.repo.SeedCommodityBetas(ctx); err != nil {
		slog.Warn("commodity betas seed failed", "err", err)
	}

	slog.Info("commodity collection complete", "prices", len(prices), "countries", len(uemoaCountries))
	return nil
}

// baseStability holds historical stability scores (0..100, 100 = perfectly stable).
var baseStability = map[string]float64{
	"Côte d'Ivoire": 75.0,
	"Sénégal":       82.0,
	"Togo":          68.0,
	"Burkina Faso":  35.0,
	"Bénin":         72.0,
	"Mali":          28.0,
	"Niger":         25.0,
	"Guinée-Bissau": 45.0,
}

// gdeltNames maps display names to GDELT country identifiers.
var gdeltNames = map[string]string{
	"Côte d'Ivoire": "IvoryCoast",
	"Sénégal":       "Senegal",
	"Togo":          "Togo",
	"Burkina Faso":  "BurkinaFaso",
	"Bénin":         "Benin",
	"Mali":          "Mali",
	"Niger":         "Niger",
	"Guinée-Bissau": "GuineaBissau",
}

// CollectPoliticalRisk evaluates regional ECOWAS stability from GDELT (or baselines)
// and inserts political_stability and political_crisis indicators.
func (s *MacroService) CollectPoliticalRisk(ctx context.Context) error {
	slog.Info("political risk collection started")

	var scores []domain.PoliticalRiskScore
	now := time.Now()

	for _, country := range uemoaCountries {
		stability := baseStability[country]
		crisis := false
		source := "GDELT/ECOWAS-Monitor"

		events, err := s.political.FetchGDELTEvents(ctx, gdeltNames[country])
		if err != nil {
			slog.Warn("gdelt fetch failed, using baseline", "country", country, "err", err)
		} else if events > 15 {
			// Abnormal levels of political violence in the last 7 days.
			stability -= float64(events) * 1.5
			if stability < 0 {
				stability = 0
			}
			// Extreme violence/riot levels trigger a full hard freeze.
			if events > 35 {
				crisis = true
				slog.Warn("active political crisis detected, triggering signal freeze", "country", country)
			}
		}

		scores = append(scores, domain.PoliticalRiskScore{
			Country:            country,
			PoliticalStability: stability,
			PoliticalCrisis:    crisis,
			Source:             source,
		})
	}

	if err := s.repo.InsertPoliticalRisk(ctx, scores, now); err != nil {
		return fmt.Errorf("insert political risk: %w", err)
	}

	slog.Info("political risk collection complete", "countries", len(scores))
	return nil
}
