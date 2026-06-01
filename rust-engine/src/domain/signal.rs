// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

use std::collections::HashMap;
use std::fmt;
use super::{Ticker, TechnicalIndicators};

#[derive(Debug, Clone)]
pub enum SignalType {
    StrongBuy,
    Buy,
    Hold,
    Sell,
    StrongSell,
}

impl fmt::Display for SignalType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            SignalType::StrongBuy => write!(f, "STRONG_BUY"),
            SignalType::Buy => write!(f, "BUY"),
            SignalType::Hold => write!(f, "HOLD"),
            SignalType::Sell => write!(f, "SELL"),
            SignalType::StrongSell => write!(f, "STRONG_SELL"),
        }
    }
}

#[derive(Debug, Clone)]
pub struct ScoringResult {
    pub ticker: Ticker,
    pub composite_score: f64,
    pub signal: SignalType,
    pub confidence: f64,
    pub reasons: Vec<String>,
    pub technical_indicators: TechnicalIndicators,
    pub fundamental_scores: HashMap<String, f64>,
    pub technical_score: f64,
    pub risk_score: f64,
    pub macro_score: f64,
    pub diversification_score: f64,
    pub macro_reasons: Vec<String>,
    /// Seasonal factor based on UEMOA harvest calendar and ex-dividend effect.
    /// 0.8 = seasonal headwind, 1.0 = neutral, 1.2 = seasonal tailwind.
    pub seasonality_score: f64,
    /// Half-Kelly fraction: optimal capital allocation [0.0, 0.25].
    /// 0.0 = no position, 0.25 = max 25% of available capital.
    pub kelly_fraction: f64,
    /// Suggested position as % of portfolio (half-Kelly, capped at 25%).
    pub suggested_position_pct: f64,
}
