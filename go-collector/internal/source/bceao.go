// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// BCEAOSource fetches macroeconomic indicators from World Bank, BRVM and ER-API.
type BCEAOSource struct {
	http *HTTPClient
}

// NewBCEAOSource creates a BCEAOSource.
func NewBCEAOSource(http *HTTPClient) *BCEAOSource {
	return &BCEAOSource{http: http}
}

var tauxDirecteurRE = regexp.MustCompile(`taux directeur[^0-9]*(\d+[,.]\d+)`)

// FetchWorldBankInflation fetches the annual inflation rate from the World Bank API.
func (s *BCEAOSource) FetchWorldBankInflation(ctx context.Context, countryCode string) (float64, error) {
	url := fmt.Sprintf(WorldBankInflationURL, countryCode)

	resp, err := s.http.Get(ctx, url)
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

// ScrapeTauxDirecteur scrapes the BCEAO taux directeur from the BRVM site.
// Returns an error if the page cannot be fetched or no rate is found, so the
// caller can apply its own fallback chain.
func (s *BCEAOSource) ScrapeTauxDirecteur(ctx context.Context) (float64, error) {
	resp, err := s.http.Get(ctx, BRVMTauxDirecteurURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Look for rate like "3.50%" or "3,50%" or "3.75".
	matches := tauxDirecteurRE.FindSubmatch(body)
	if len(matches) > 1 {
		valStr := strings.ReplaceAll(string(matches[1]), ",", ".")
		if val, err := strconv.ParseFloat(valStr, 64); err == nil {
			return val, nil
		}
	}

	return 0, fmt.Errorf("taux directeur not found on page")
}

// FetchXOFUSDRate fetches the real-time XOF/USD exchange rate.
func (s *BCEAOSource) FetchXOFUSDRate(ctx context.Context) (float64, error) {
	resp, err := s.http.Get(ctx, ERAPIExchangeRateURL)
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
