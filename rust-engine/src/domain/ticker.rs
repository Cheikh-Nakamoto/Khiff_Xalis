// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

#[derive(Debug, Clone)]
pub struct Ticker {
    pub symbol: String,
    pub name: String,
    pub country: String,
    pub sector: String,
}

#[derive(Debug, Clone)]
pub struct PortfolioPosition {
    pub ticker: String,
    pub country: String,
    pub sector: String,
    pub market_value: f64,
}
