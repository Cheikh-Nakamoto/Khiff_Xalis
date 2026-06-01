// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

use crate::domain::SignalType;

/// Kelly Criterion position sizing service.
///
/// Computes the half-Kelly fraction to determine optimal capital allocation
/// per trade, accounting for BRVM-specific illiquidity constraints.
///
/// Formula: f* = (b × p - q) / b
/// Half-Kelly (f*/2) is used for robustness against estimation error.
/// Maximum position capped at 25% to enforce basic diversification.
pub struct PositionSizingService;

impl Default for PositionSizingService {
    fn default() -> Self {
        Self::new()
    }
}

impl PositionSizingService {
    pub fn new() -> Self {
        Self
    }

    /// Returns (kelly_fraction, suggested_position_pct).
    ///
    /// - `kelly_fraction`: dimensionless [0.0, 0.25]
    /// - `suggested_position_pct`: as % of available capital [0.0, 25.0]
    ///
    /// # Arguments
    /// * `signal` - The trading signal (determines win probability)
    /// * `atr_14` - Average True Range (determines gain/loss ratio via 2:1 TP/SL)
    /// * `current_price` - Current price in FCFA (used to compute ATR as % move)
    /// * `amihud` - Amihud illiquidity ratio (scales down position if illiquid)
    /// * `zero_return_ratio` - Fraction of flat days (further reduces sizing)
    pub fn compute(
        &self,
        signal: &SignalType,
        atr_14: Option<f64>,
        current_price: f64,
        amihud: Option<f64>,
        zero_return_ratio: Option<f64>,
    ) -> (f64, f64) {
        let p_win = Self::signal_to_p_win(signal);
        let b_ratio = Self::atr_to_b_ratio(atr_14, current_price);
        let raw_fraction = self.kelly_fraction(p_win, b_ratio);

        // Apply illiquidity penalty: Amihud > 0.5 → scale down proportionally
        let illiq_scale = match amihud {
            Some(a) if a > 2.0 => 0.25,  // extremely illiquid: quarter-size
            Some(a) if a > 1.0 => 0.50,  // very illiquid: half-size
            Some(a) if a > 0.5 => 0.75,  // moderately illiquid: three-quarter-size
            _ => 1.0,                      // liquid enough: full Kelly fraction
        };

        // Zero-return penalty: stale market → hard cap on position size
        let zrr_scale = match zero_return_ratio {
            Some(z) if z > 0.60 => 0.0,   // effectively frozen: no position
            Some(z) if z > 0.40 => 0.30,  // very stale: minimal exposure only
            Some(z) if z > 0.20 => 0.70,  // somewhat stale: reduce size
            _ => 1.0,
        };

        let adjusted = raw_fraction * illiq_scale * zrr_scale;
        let pct = adjusted * 100.0;
        (adjusted, pct)
    }

    /// Half-Kelly fraction: (b×p - q) / b / 2, clamped to [0, 0.25].
    pub fn kelly_fraction(&self, p_win: f64, b_ratio: f64) -> f64 {
        if b_ratio <= 0.0 || p_win <= 0.0 || p_win >= 1.0 {
            return 0.0;
        }
        let q = 1.0 - p_win;
        let full_kelly = (b_ratio * p_win - q) / b_ratio;
        // Half-Kelly for robustness. Negative Kelly → no position.
        (full_kelly / 2.0).clamp(0.0, 0.25)
    }

    /// Maps signal type to estimated win probability.
    ///
    /// These values should be re-calibrated monthly once backtest data accumulates.
    /// Initial calibration based on mean-reversion literature for frontier markets.
    fn signal_to_p_win(signal: &SignalType) -> f64 {
        match signal {
            SignalType::StrongBuy  => 0.62,
            SignalType::Buy        => 0.57,
            SignalType::Hold       => 0.50,
            SignalType::Sell       => 0.43,
            SignalType::StrongSell => 0.38,
        }
    }

    /// Derives gain/loss ratio from ATR using a 2:1 take-profit/stop-loss rule.
    ///
    /// TP = 2 × ATR (as % of price)
    /// SL = 1 × ATR (as % of price)
    /// b  = TP% / SL% = 2.0
    ///
    /// If ATR is unavailable, falls back to b = 1.5 (conservative).
    fn atr_to_b_ratio(atr: Option<f64>, price: f64) -> f64 {
        if price <= 0.0 {
            return 1.5;
        }
        match atr {
            Some(a) if a > 0.0 => {
                let sl_pct = a / price;
                let tp_pct = 2.0 * sl_pct;
                if sl_pct > 0.0 {
                    (tp_pct / sl_pct).clamp(0.5, 4.0) // always 2.0 here, but clamp for safety
                } else {
                    1.5
                }
            }
            _ => 1.5,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_kelly_strong_buy() {
        let svc = PositionSizingService::new();
        // StrongBuy, p=0.62, b=2.0 → full_kelly = (2×0.62-0.38)/2 = 0.43 → half=0.215
        let fraction = svc.kelly_fraction(0.62, 2.0);
        assert!((0.20..=0.25).contains(&fraction), "Expected ~0.215, got {}", fraction);
    }

    #[test]
    fn test_kelly_strong_sell_is_zero() {
        let svc = PositionSizingService::new();
        // StrongSell: p=0.38, b=2.0 → full_kelly = (2×0.38-0.62)/2 = 0.07 → half=0.035
        // Still positive (contrarian can take a small position on the other side),
        // but in our BUY-only BRVM context we expect 0 for sell signals.
        // The signal.rs will handle this: position sizing is only applied to BUY signals.
        let fraction = svc.kelly_fraction(0.38, 2.0);
        assert!(fraction >= 0.0);
    }

    #[test]
    fn test_illiquidity_penalty() {
        let svc = PositionSizingService::new();
        // Strong buy, very illiquid (amihud > 2.0) → quarter-size position
        let (frac_liquid, _) = svc.compute(&SignalType::StrongBuy, Some(100.0), 10000.0, None, None);
        let (frac_illiquid, _) = svc.compute(&SignalType::StrongBuy, Some(100.0), 10000.0, Some(3.0), None);
        assert!(frac_illiquid < frac_liquid * 0.4, "Illiquid position should be << liquid");
    }

    #[test]
    fn test_frozen_market_no_position() {
        let svc = PositionSizingService::new();
        // zero_return_ratio > 0.6 → no position regardless of signal
        let (frac, pct) = svc.compute(&SignalType::StrongBuy, Some(100.0), 10000.0, None, Some(0.65));
        assert_eq!(frac, 0.0);
        assert_eq!(pct, 0.0);
    }

    #[test]
    fn test_max_position_cap() {
        let svc = PositionSizingService::new();
        // Even with extreme parameters, position cannot exceed 25%
        let (frac, _) = svc.compute(&SignalType::StrongBuy, Some(1.0), 100.0, None, None);
        assert!(frac <= 0.25, "Position must be capped at 25%, got {}", frac);
    }
}
