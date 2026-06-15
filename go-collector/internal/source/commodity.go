// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package source

import (
	"context"
	"encoding/json"
	"fmt"
)

// CommoditySource fetches commodity spot prices from ICCO and Trading Economics.
type CommoditySource struct {
	http *HTTPClient
}

// NewCommoditySource creates a CommoditySource.
func NewCommoditySource(http *HTTPClient) *CommoditySource {
	return &CommoditySource{http: http}
}

// FetchCocoaPrice fetches the daily cocoa price from ICCO (USD/tonne).
func (s *CommoditySource) FetchCocoaPrice(ctx context.Context) (float64, error) {
	resp, err := s.http.Get(ctx, ICCOCocoaPriceURL)
	if err != nil {
		return 0, fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Price float64 `json:"price"`
			Date  string  `json:"date"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode JSON: %w", err)
	}

	if len(result.Data) == 0 {
		return 0, fmt.Errorf("no cocoa price data available")
	}

	return result.Data[0].Price, nil
}

// FetchOilPrice fetches the Brent crude oil price.
func (s *CommoditySource) FetchOilPrice(ctx context.Context) (float64, error) {
	return s.FetchFromTradingEconomics(ctx, "Brent")
}

// FetchFromTradingEconomics pulls a generic commodity from the Trading Economics
// guest API feed.
func (s *CommoditySource) FetchFromTradingEconomics(ctx context.Context, name string) (float64, error) {
	resp, err := s.http.Get(ctx, TradingEconomicsCommoditiesURL)
	if err != nil {
		return 0, fmt.Errorf("HTTP request: %w", err)
	}
	defer resp.Body.Close()

	var result []struct {
		Name string  `json:"Name"`
		Last float64 `json:"Last"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode JSON: %w", err)
	}

	for _, item := range result {
		if item.Name == name || (name == "Brent" && (item.Name == "Crude Oil" || item.Name == "Brent Crude")) {
			return item.Last, nil
		}
	}

	return 0, fmt.Errorf("commodity %s not found in response", name)
}
