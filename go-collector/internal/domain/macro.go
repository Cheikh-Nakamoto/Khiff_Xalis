// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package domain

import "time"

// BCEAOIndicator represents a macroeconomic indicator from BCEAO.
type BCEAOIndicator struct {
	Country   string
	Indicator string
	Value     float64
	Unit      string
	Date      time.Time
}

// CommodityPrice represents a commodity price data point.
type CommodityPrice struct {
	Commodity string
	Price     float64
	Unit      string
	Source    string
	Date      time.Time
}

// PoliticalRiskScore represents evaluated regional risk.
type PoliticalRiskScore struct {
	Country            string
	PoliticalStability float64
	PoliticalCrisis    bool
	Source             string
}
