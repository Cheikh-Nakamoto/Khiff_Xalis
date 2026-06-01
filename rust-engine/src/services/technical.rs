// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

use crate::domain::TechnicalIndicators;

pub struct TechnicalAnalysisService;

impl Default for TechnicalAnalysisService {
    fn default() -> Self {
        Self::new()
    }
}

impl TechnicalAnalysisService {
    pub fn new() -> Self {
        Self
    }

    pub fn calculate(&self, closes: &[f64], volumes: &[i64], highs: &[f64], lows: &[f64]) -> TechnicalIndicators {
        let mut indicators = TechnicalIndicators::default();

        if closes.len() < 50 {
            return indicators;
        }

        indicators.sma_20 = Some(self.sma(closes, 20));
        indicators.sma_50 = Some(self.sma(closes, 50));
        indicators.ema_12 = Some(self.ema(closes, 12));
        indicators.ema_26 = Some(self.ema(closes, 26));

        if let (Some(ema12), Some(ema26)) = (indicators.ema_12, indicators.ema_26) {
            indicators.macd = Some(ema12 - ema26);
            let macd_values: Vec<f64> = (0..9)
                .map(|i| {
                    let end = closes.len() - i;
                    let e12 = self.ema(&closes[..end], 12);
                    let e26 = self.ema(&closes[..end], 26);
                    e12 - e26
                })
                .collect();
            indicators.macd_signal = Some(self.ema(&macd_values, 9));
        }

        indicators.rsi_14 = Some(self.rsi(closes, 14));

        if closes.len() >= 20 {
            let sma20 = self.sma(closes, 20);
            let std20 = self.std_dev(&closes[closes.len()-20..], sma20);
            indicators.bollinger_upper = Some(sma20 + 2.0 * std20);
            indicators.bollinger_lower = Some(sma20 - 2.0 * std20);
        }

        indicators.atr_14 = Some(self.atr(highs, lows, closes, 14));

        if volumes.len() >= 20 {
            let sum: i64 = volumes[volumes.len()-20..].iter().sum();
            indicators.volume_sma_20 = Some(sum as f64 / 20.0);
        }

        indicators
    }

    pub fn sma(&self, prices: &[f64], period: usize) -> f64 {
        let slice = &prices[prices.len().saturating_sub(period)..];
        slice.iter().sum::<f64>() / slice.len() as f64
    }

    pub fn ema(&self, prices: &[f64], period: usize) -> f64 {
        if prices.len() < period {
            return prices.last().copied().unwrap_or(0.0);
        }
        let k = 2.0 / (period as f64 + 1.0);
        let mut ema = self.sma(prices, period);
        for &price in &prices[period..] {
            ema = price * k + ema * (1.0 - k);
        }
        ema
    }

    pub fn rsi(&self, prices: &[f64], period: usize) -> f64 {
        if prices.len() <= period {
            return 50.0;
        }

        let mut avg_gain = 0.0;
        let mut avg_loss = 0.0;

        // Initial SMA for the first 'period' elements
        for i in 1..=period {
            let diff = prices[i] - prices[i - 1];
            if diff > 0.0 {
                avg_gain += diff;
            } else {
                avg_loss += diff.abs();
            }
        }

        avg_gain /= period as f64;
        avg_loss /= period as f64;

        // Wilder's Smoothing for the rest of the data
        for i in (period + 1)..prices.len() {
            let diff = prices[i] - prices[i - 1];
            let gain = if diff > 0.0 { diff } else { 0.0 };
            let loss = if diff < 0.0 { diff.abs() } else { 0.0 };

            avg_gain = (avg_gain * (period as f64 - 1.0) + gain) / period as f64;
            avg_loss = (avg_loss * (period as f64 - 1.0) + loss) / period as f64;
        }

        if avg_loss == 0.0 {
            if avg_gain == 0.0 {
                return 50.0;
            }
            return 100.0;
        }

        let rs = avg_gain / avg_loss;
        100.0 - (100.0 / (1.0 + rs))
    }

    pub fn std_dev(&self, prices: &[f64], mean: f64) -> f64 {
        let variance = prices.iter().map(|p| (p - mean).powi(2)).sum::<f64>() / prices.len() as f64;
        variance.sqrt()
    }

    pub fn atr(&self, highs: &[f64], lows: &[f64], closes: &[f64], period: usize) -> f64 {
        if highs.len() < period + 1 || lows.len() < period + 1 || closes.len() < period + 1 {
            return 0.0;
        }
        let mut tr_sum = 0.0;
        for i in 1..=period {
            let idx = closes.len() - i;
            let prev_idx = idx - 1;
            let high = highs[idx];
            let low = lows[idx];
            let prev_close = closes[prev_idx];
            let tr = (high - low)
                .max((high - prev_close).abs())
                .max((low - prev_close).abs());
            tr_sum += tr;
        }
        tr_sum / period as f64
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_sma() {
        let service = TechnicalAnalysisService::new();
        let prices = vec![10.0, 11.0, 12.0, 13.0, 14.0, 15.0, 16.0, 17.0, 18.0, 19.0,
                          20.0, 21.0, 22.0, 23.0, 24.0, 25.0, 26.0, 27.0, 28.0, 29.0,
                          30.0, 31.0, 32.0, 33.0, 34.0, 35.0, 36.0, 37.0, 38.0, 39.0,
                          40.0, 41.0, 42.0, 43.0, 44.0, 45.0, 46.0, 47.0, 48.0, 49.0,
                          50.0, 51.0, 52.0, 53.0, 54.0, 55.0, 56.0, 57.0, 58.0, 59.0,
                          60.0, 61.0, 62.0, 63.0, 64.0, 65.0, 66.0, 67.0, 68.0, 69.0];
        let sma = service.sma(&prices, 20);
        assert!((sma - 59.5).abs() < 0.01);
    }

    #[test]
    fn test_rsi() {
        let service = TechnicalAnalysisService::new();
        let prices_up: Vec<f64> = (0..20).map(|i| 100.0 + i as f64).collect();
        let rsi = service.rsi(&prices_up, 14);
        assert!(rsi > 70.0);

        let prices_down: Vec<f64> = (0..20).map(|i| 100.0 - i as f64).collect();
        let rsi = service.rsi(&prices_down, 14);
        assert!(rsi < 30.0);
    }
}
