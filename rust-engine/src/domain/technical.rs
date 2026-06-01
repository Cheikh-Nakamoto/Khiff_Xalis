// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

#[derive(Debug, Clone, Default)]
pub struct TechnicalIndicators {
    pub sma_20: Option<f64>,
    pub sma_50: Option<f64>,
    pub ema_12: Option<f64>,
    pub ema_26: Option<f64>,
    pub rsi_14: Option<f64>,
    pub macd: Option<f64>,
    pub macd_signal: Option<f64>,
    pub bollinger_upper: Option<f64>,
    pub bollinger_lower: Option<f64>,
    pub atr_14: Option<f64>,
    pub volume_sma_20: Option<f64>,
    /// Amihud illiquidity ratio over 20 days.
    /// ILLIQ = (1/D) × Σ |r_d| / (volume_d × price_d) × 10^6
    /// Higher = more illiquid. None if insufficient data.
    pub amihud_20: Option<f64>,
    /// Fraction of days in the last 20 with zero return (price unchanged).
    /// > 0.4 = dangerously illiquid, signal should be suppressed.
    pub zero_return_ratio: Option<f64>,
}
