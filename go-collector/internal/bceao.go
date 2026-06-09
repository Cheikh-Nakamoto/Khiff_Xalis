// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// BCEAOIndicator represents a macroeconomic indicator from BCEAO.
type BCEAOIndicator struct {
	Country   string
	Indicator string
	Value     float64
	Unit      string
	Date      time.Time
}

// CollectBCEAO fetches macroeconomic indicators (inflation, taux directeur, change XOF/EUR, change XOF/USD)
// from BCEAO/World Bank/ECB and inserts them into the macroeconomic_data table.
func (c *Collector) CollectBCEAO(ctx context.Context) error {
	log.Println("[BCEAO] Starting macroeconomic data collection...")

	var indicators []BCEAOIndicator
	now := time.Now()

	// UEMOA member countries
	countries := []string{"CI", "SN", "TG", "BF", "BJ", "ML", "NE", "GW"}
	countryNames := map[string]string{
		"CI": "Côte d'Ivoire",
		"SN": "Sénégal",
		"TG": "Togo",
		"BF": "Burkina Faso",
		"BJ": "Bénin",
		"ML": "Mali",
		"NE": "Niger",
		"GW": "Guinée-Bissau",
	}

	// 1. Fetch inflation data from World Bank API
	for _, code := range countries {
		inflation, err := c.fetchWorldBankInflation(ctx, code)
		if err != nil {
			log.Printf("[BCEAO] Warning: could not fetch inflation for %s: %v", code, err)
			continue
		}
		indicators = append(indicators, BCEAOIndicator{
			Country:   countryNames[code],
			Indicator: "inflation",
			Value:     inflation,
			Unit:      "%",
			Date:      now,
		})
	}

	// 2. BCEAO taux directeur (scraped or fallbacks)
	tauxDirecteur, err := c.fetchBCEAORate(ctx)
	if err != nil {
		log.Printf("[BCEAO] Warning: could not fetch taux directeur: %v, using default 3.50", err)
		tauxDirecteur = 3.50
	}
	for _, code := range countries {
		indicators = append(indicators, BCEAOIndicator{
			Country:   countryNames[code],
			Indicator: "taux_directeur",
			Value:     tauxDirecteur,
			Unit:      "%",
			Date:      now,
		})
	}

	// 3. XOF/EUR exchange rate (fixed peg at 655.957)
	xofEurRate := 655.957
	for _, code := range countries {
		indicators = append(indicators, BCEAOIndicator{
			Country:   countryNames[code],
			Indicator: "change_xof_eur",
			Value:     xofEurRate,
			Unit:      "FCFA/EUR",
			Date:      now,
		})
	}

	// 4. XOF/USD exchange rate (real-time ECB/ExchangeRate API)
	xofUsdRate, err := c.fetchXOFUSDRate(ctx)
	if err != nil {
		log.Printf("[BCEAO] Warning: could not fetch XOF/USD exchange rate: %v, falling back to 600.0", err)
		xofUsdRate = 600.0 // Reasonable historic peg-based USD value
	}
	for _, code := range countries {
		indicators = append(indicators, BCEAOIndicator{
			Country:   countryNames[code],
			Indicator: "change_xof_usd",
			Value:     xofUsdRate,
			Unit:      "FCFA/USD",
			Date:      now,
		})
	}

	// Batch insert
	if err := c.insertMacroData(ctx, indicators); err != nil {
		return fmt.Errorf("insert macro data: %w", err)
	}

	log.Printf("[BCEAO] Collected %d macroeconomic indicators", len(indicators))
	return nil
}

// fetchWorldBankInflation fetches annual inflation rate from World Bank API.
func (c *Collector) fetchWorldBankInflation(ctx context.Context, countryCode string) (float64, error) {
	url := fmt.Sprintf(WorldBankInflationURL, countryCode)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	var result []json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode JSON: %w", err)
	}

	if len(result) < 2 {
		return 0, fmt.Errorf("unexpected response format")
	}

	var data []struct {
		Value *float64 `json:"value"`
		Date  string   `json:"date"`
	}
	if err := json.Unmarshal(result[1], &data); err != nil {
		return 0, fmt.Errorf("parse data: %w", err)
	}

	for _, d := range data {
		if d.Value != nil {
			return *d.Value, nil
		}
	}

	return 0, fmt.Errorf("no inflation data for %s", countryCode)
}

// fetchBCEAORate attempts to scrape the BCEAO taux directeur.
// Falls back to last known rate in DB, or 3.50 if database is empty.
func (c *Collector) fetchBCEAORate(ctx context.Context) (float64, error) {
	// First, fetch the last stored rate in our database as the fallback
	var lastRate float64
	dbErr := c.DB.QueryRow(ctx,
		`SELECT value FROM macroeconomic_data 
		 WHERE indicator = 'taux_directeur' 
		 ORDER BY time DESC LIMIT 1`).Scan(&lastRate)

	fallbackRate := 3.50
	if dbErr == nil && lastRate > 0 {
		fallbackRate = lastRate
	}

	// Try scraping BRVM/BCEAO news
	req, err := http.NewRequestWithContext(ctx, "GET", BRVMTauxDirecteurURL, nil)
	if err != nil {
		return fallbackRate, nil
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fallbackRate, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fallbackRate, nil
	}

	// Look for rate like "3.50%" or "3,50%" or "3.75"
	re := regexp.MustCompile(`taux directeur[^0-9]*(\d+[,.]\d+)`)
	matches := re.FindSubmatch(body)
	if len(matches) > 1 {
		valStr := strings.ReplaceAll(string(matches[1]), ",", ".")
		if val, err := strconv.ParseFloat(valStr, 64); err == nil {
			return val, nil
		}
	}

	return fallbackRate, nil
}

// fetchXOFUSDRate fetches the real-time XOF/USD exchange rate.
func (c *Collector) fetchXOFUSDRate(ctx context.Context) (float64, error) {
	url := ERAPIExchangeRateURL
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Result string             `json:"result"`
		Rates  map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if result.Result != "success" {
		return 0, fmt.Errorf("API error status: %s", result.Result)
	}

	rate, ok := result.Rates["XOF"]
	if !ok {
		return 0, fmt.Errorf("XOF rate not found in USD rates response")
	}

	return rate, nil
}

// insertMacroData inserts macroeconomic indicators into the database.
func (c *Collector) insertMacroData(ctx context.Context, indicators []BCEAOIndicator) error {
	batch := &pgx.Batch{}

	for _, ind := range indicators {
		batch.Queue(
			`INSERT INTO macroeconomic_data (time, country, indicator, value, unit, source, confidence)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			ind.Date, ind.Country, ind.Indicator, ind.Value, ind.Unit, "BCEAO/WorldBank", 0.9,
		)
	}

	br := c.DB.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert row %d: %w", i, err)
		}
	}

	return nil
}
