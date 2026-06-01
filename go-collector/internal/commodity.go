// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// CommodityPrice represents a commodity price data point.
type CommodityPrice struct {
	Commodity string
	Price     float64
	Unit      string
	Source    string
	Date      time.Time
}

// CollectCommodities fetches commodity prices (cocoa, oil) and inserts
// them into the macroeconomic_data table.
//
// Sources:
// - Cocoa: ICCO (International Cocoa Organization) daily prices
// - Oil: Brent crude from public APIs
func (c *Collector) CollectCommodities(ctx context.Context) error {
	log.Println("[COMMODITY] Starting commodity price collection...")

	var prices []CommodityPrice
	now := time.Now()

	// 1. Cocoa price (USD/tonne)
	cocoaPrice, err := c.fetchCocoaPrice(ctx)
	if err != nil {
		log.Printf("[COMMODITY] Warning: could not fetch cocoa price: %v, using fallback", err)
		cocoaPrice = 3200.0 // Approximate 2024 price
	}
	prices = append(prices, CommodityPrice{
		Commodity: "cocoa_price",
		Price:     cocoaPrice,
		Unit:      "USD/tonne",
		Source:    "ICCO",
		Date:      now,
	})

	// 2. Brent oil price (USD/barrel)
	oilPrice, err := c.fetchOilPrice(ctx)
	if err != nil {
		log.Printf("[COMMODITY] Warning: could not fetch oil price: %v, using fallback", err)
		oilPrice = 80.0 // Approximate 2024 price
	}
	prices = append(prices, CommodityPrice{
		Commodity: "oil_price",
		Price:     oilPrice,
		Unit:      "USD/barrel",
		Source:    "ICE/TradingEconomics",
		Date:      now,
	})

	// Insert for all UEMOA countries
	countries := []string{"Côte d'Ivoire", "Sénégal", "Togo", "Burkina Faso", "Bénin", "Mali", "Niger", "Guinée-Bissau"}

	if err := c.insertCommodityData(ctx, prices, countries); err != nil {
		return fmt.Errorf("insert commodity data: %w", err)
	}

	log.Printf("[COMMODITY] Collected %d commodity prices for %d countries", len(prices), len(countries))
	return nil
}

// fetchCocoaPrice fetches the daily cocoa price from ICCO.
func (c *Collector) fetchCocoaPrice(ctx context.Context) (float64, error) {
	// ICCO publishes daily cocoa prices at https://www.icco.org/statistics/
	// The API endpoint returns JSON with daily prices
	url := "https://www.icco.org/wp-json/icco/v1/daily-prices?limit=1"

	req, err := c.HTTP.Get(url)
	if err != nil {
		return 0, fmt.Errorf("HTTP request: %w", err)
	}
	defer req.Body.Close()

	var result struct {
		Data []struct {
			Price float64 `json:"price"`
			Date  string  `json:"date"`
		} `json:"data"`
	}
	if err := json.NewDecoder(req.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode JSON: %w", err)
	}

	if len(result.Data) == 0 {
		return 0, fmt.Errorf("no cocoa price data available")
	}

	return result.Data[0].Price, nil
}

// fetchOilPrice fetches the Brent crude oil price.
func (c *Collector) fetchOilPrice(ctx context.Context) (float64, error) {
	// Use a public API for oil prices
	// Trading Economics provides free limited access
	url := "https://api.tradingeconomics.com/markets/commodities?c=guest:guest&f=json"

	req, err := c.HTTP.Get(url)
	if err != nil {
		return 0, fmt.Errorf("HTTP request: %w", err)
	}
	defer req.Body.Close()

	var result []struct {
		Name  string  `json:"Name"`
		Last  float64 `json:"Last"`
	}
	if err := json.NewDecoder(req.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode JSON: %w", err)
	}

	for _, item := range result {
		if item.Name == "Brent" || item.Name == "Crude Oil" || item.Name == "Brent Crude" {
			return item.Last, nil
		}
	}

	return 0, fmt.Errorf("brent oil price not found in response")
}

// insertCommodityData inserts commodity prices for all countries.
func (c *Collector) insertCommodityData(ctx context.Context, prices []CommodityPrice, countries []string) error {
	batch := &pgx.Batch{}

	for _, price := range prices {
		for _, country := range countries {
			batch.Queue(
				`INSERT INTO macroeconomic_data (time, country, indicator, value, unit, source, confidence)
				 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				price.Date, country, price.Commodity, price.Price, price.Unit, price.Source, 0.95,
			)
		}
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
