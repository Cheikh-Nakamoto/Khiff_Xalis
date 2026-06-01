// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

#[derive(Debug, Clone, Default)]
pub struct FundamentalData {
    pub per: Option<f64>,
    pub roe: Option<f64>,
    pub dividend_yield: Option<f64>,
    pub eps: Option<f64>,
    pub book_value_per_share: Option<f64>,
    pub debt_to_equity: Option<f64>,
    pub revenue_growth: Option<f64>,
    pub net_profit: Option<f64>,
}
