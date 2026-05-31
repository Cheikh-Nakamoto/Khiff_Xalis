// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

use std::collections::HashMap;

/// ============================================================
/// BRVM RUST ENGINE - Core Calculations
/// Technical Analysis + Fundamental Scoring + Signal Generation
/// ============================================================

// ─────────────────────────────────────────
// DOMAIN: Entities
// ─────────────────────────────────────────

#[derive(Debug, Clone)]
pub struct Ticker {
    pub symbol: String,
    pub name: String,
    pub country: String,
    pub sector: String,
}

#[derive(Debug, Clone)]
pub struct MarketDataPoint {
    pub date: String,
    pub open: f64,
    pub high: f64,
    pub low: f64,
    pub close: f64,
    pub volume: i64,
}

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
}

#[derive(Debug, Clone)]
pub enum SignalType {
    StrongBuy,
    Buy,
    Hold,
    Sell,
    StrongSell,
}

impl SignalType {
    pub fn to_string(&self) -> String {
        match self {
            SignalType::StrongBuy => "STRONG_BUY".to_string(),
            SignalType::Buy => "BUY".to_string(),
            SignalType::Hold => "HOLD".to_string(),
            SignalType::Sell => "SELL".to_string(),
            SignalType::StrongSell => "STRONG_SELL".to_string(),
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
}

// ─────────────────────────────────────────
// USE CASE: Technical Analysis
// ─────────────────────────────────────────

pub struct TechnicalAnalysisService;

impl TechnicalAnalysisService {
    pub fn new() -> Self {
        Self
    }

    pub fn calculate(&self, closes: &[f64], volumes: &[i64], highs: &[f64], lows: &[f64]) -> TechnicalIndicators {
        let mut indicators = TechnicalIndicators::default();

        if closes.len() < 50 {
            return indicators;
        }

        // SMA
        indicators.sma_20 = Some(self.sma(closes, 20));
        indicators.sma_50 = Some(self.sma(closes, 50));

        // EMA
        indicators.ema_12 = Some(self.ema(closes, 12));
        indicators.ema_26 = Some(self.ema(closes, 26));

        // MACD
        if let (Some(ema12), Some(ema26)) = (indicators.ema_12, indicators.ema_26) {
            indicators.macd = Some(ema12 - ema26);
            // MACD Signal = EMA(9) of MACD
            // Simplified: we compute EMA of the last 9 MACD values
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

        // RSI
        indicators.rsi_14 = Some(self.rsi(closes, 14));

        // Bollinger Bands
        if closes.len() >= 20 {
            let sma20 = self.sma(closes, 20);
            let std20 = self.std_dev(&closes[closes.len()-20..], sma20);
            indicators.bollinger_upper = Some(sma20 + 2.0 * std20);
            indicators.bollinger_lower = Some(sma20 - 2.0 * std20);
        }

        // ATR (True Range based)
        indicators.atr_14 = Some(self.atr(highs, lows, closes, 14));

        // Volume SMA
        if volumes.len() >= 20 {
            let sum: i64 = volumes[volumes.len()-20..].iter().sum();
            indicators.volume_sma_20 = Some(sum as f64 / 20.0);
        }

        indicators
    }

    fn sma(&self, prices: &[f64], period: usize) -> f64 {
        let slice = &prices[prices.len().saturating_sub(period)..];
        slice.iter().sum::<f64>() / slice.len() as f64
    }

    fn ema(&self, prices: &[f64], period: usize) -> f64 {
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

    fn rsi(&self, prices: &[f64], period: usize) -> f64 {
        if prices.len() < period + 1 {
            return 50.0;
        }
        let mut gains = 0.0;
        let mut losses = 0.0;
        for i in 1..=period {
            let diff = prices[prices.len() - i] - prices[prices.len() - i - 1];
            if diff > 0.0 {
                gains += diff;
            } else {
                losses += diff.abs();
            }
        }
        let avg_gain = gains / period as f64;
        let avg_loss = losses / period as f64;
        if avg_loss == 0.0 {
            return 100.0;
        }
        let rs = avg_gain / avg_loss;
        100.0 - (100.0 / (1.0 + rs))
    }

    fn std_dev(&self, prices: &[f64], mean: f64) -> f64 {
        let variance = prices.iter().map(|p| (p - mean).powi(2)).sum::<f64>() / prices.len() as f64;
        variance.sqrt()
    }

    fn atr(&self, highs: &[f64], lows: &[f64], closes: &[f64], period: usize) -> f64 {
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

// ─────────────────────────────────────────
// USE CASE: Fundamental Scoring
// ─────────────────────────────────────────

pub struct FundamentalScoringService;

impl FundamentalScoringService {
    pub fn new() -> Self {
        Self
    }

    pub fn score(&self, fundamental: &FundamentalData) -> (HashMap<String, f64>, Vec<String>) {
        let mut scores = HashMap::new();
        let mut reasons = Vec::new();

        // Score PER (inversé)
        let per_score = match fundamental.per {
            Some(p) if p < 8.0 => { reasons.push(format!("PER très attractif: {:.1}", p)); 95.0 }
            Some(p) if p < 12.0 => { reasons.push(format!("PER attractif: {:.1}", p)); 80.0 }
            Some(p) if p < 18.0 => 60.0,
            Some(p) if p < 25.0 => { reasons.push(format!("PER élevé: {:.1}", p)); 40.0 }
            Some(p) => { reasons.push(format!("PER très élevé: {:.1}", p)); 20.0 }
            None => 50.0,
        };
        scores.insert("per".to_string(), per_score);

        // Score ROE
        let roe_score = match fundamental.roe {
            Some(r) if r > 25.0 => { reasons.push(format!("ROE excellent: {:.1}%", r)); 95.0 }
            Some(r) if r > 20.0 => { reasons.push(format!("ROE très bon: {:.1}%", r)); 85.0 }
            Some(r) if r > 15.0 => 70.0,
            Some(r) if r > 10.0 => 50.0,
            Some(r) => { reasons.push(format!("ROE faible: {:.1}%", r)); 30.0 }
            None => 50.0,
        };
        scores.insert("roe".to_string(), roe_score);

        // Score Dividend Yield
        let div_score = match fundamental.dividend_yield {
            Some(d) if d > 7.0 => { reasons.push(format!("Yield exceptionnel: {:.1}%", d)); 95.0 }
            Some(d) if d > 5.0 => { reasons.push(format!("Yield attractif: {:.1}%", d)); 80.0 }
            Some(d) if d > 3.0 => 60.0,
            Some(d) if d > 1.0 => 40.0,
            Some(_) => 20.0,
            None => { reasons.push("Pas de dividende".to_string()); 30.0 }
        };
        scores.insert("dividend".to_string(), div_score);

        // Score Dette
        let debt_score = match fundamental.debt_to_equity {
            Some(d) if d < 0.5 => 90.0,
            Some(d) if d < 1.0 => 70.0,
            Some(d) if d < 1.5 => 50.0,
            Some(d) => { reasons.push(format!("Dette élevée: {:.2}", d)); 30.0 }
            None => 50.0,
        };
        scores.insert("debt".to_string(), debt_score);

        // Score Croissance
        let growth_score = match fundamental.revenue_growth {
            Some(g) if g > 20.0 => { reasons.push(format!("Croissance forte: {:.1}%", g)); 90.0 }
            Some(g) if g > 10.0 => 75.0,
            Some(g) if g > 0.0 => 60.0,
            Some(g) => { reasons.push(format!("Croissance négative: {:.1}%", g)); 30.0 }
            None => 50.0,
        };
        scores.insert("growth".to_string(), growth_score);

        (scores, reasons)
    }
}

// ─────────────────────────────────────────
// USE CASE: Signal Generation
// ─────────────────────────────────────────

pub struct SignalGenerationService {
    tech: TechnicalAnalysisService,
    fund: FundamentalScoringService,
}

impl SignalGenerationService {
    pub fn new() -> Self {
        Self {
            tech: TechnicalAnalysisService::new(),
            fund: FundamentalScoringService::new(),
        }
    }

    pub fn generate_signal(
        &self,
        ticker: Ticker,
        closes: &[f64],
        volumes: &[i64],
        highs: &[f64],
        lows: &[f64],
        fundamental: &FundamentalData,
        current_price: f64,
    ) -> ScoringResult {
        // 1. Indicateurs techniques
        let tech_indicators = self.tech.calculate(closes, volumes, highs, lows);

        // 2. Score fondamental
        let (fund_scores, mut reasons) = self.fund.score(fundamental);

        // 3. Score technique
        let tech_score = self.score_technical(&tech_indicators, current_price, closes, volumes);

        // 4. Score risque
        let risk_score = self.score_risk(&tech_indicators, volumes);

        // 5. Composite
        let composite =
            fund_scores.get("per").unwrap_or(&50.0) * 0.20 +
            fund_scores.get("roe").unwrap_or(&50.0) * 0.20 +
            fund_scores.get("dividend").unwrap_or(&50.0) * 0.15 +
            fund_scores.get("growth").unwrap_or(&50.0) * 0.10 +
            fund_scores.get("debt").unwrap_or(&50.0) * 0.05 +
            tech_score * 0.20 +
            risk_score * 0.10;

        // 6. Signal
        let (signal, confidence) = self.determine_signal(composite, &tech_indicators, &fund_scores, &mut reasons);

        ScoringResult {
            ticker,
            composite_score: composite,
            signal,
            confidence,
            reasons,
            technical_indicators: tech_indicators,
            fundamental_scores: fund_scores,
            technical_score: tech_score,
            risk_score,
        }
    }

    fn score_technical(&self, tech: &TechnicalIndicators, price: f64, _prices: &[f64], volumes: &[i64]) -> f64 {
        let mut score: f64 = 50.0;

        // RSI
        if let Some(rsi) = tech.rsi_14 {
            if rsi < 30.0 {
                score += 20.0;
            } else if rsi > 70.0 {
                score -= 20.0;
            } else if rsi >= 40.0 && rsi <= 60.0 {
                score += 5.0;
            }
        }

        // Trend (SMA)
        if let (Some(sma20), Some(sma50)) = (tech.sma_20, tech.sma_50) {
            if sma20 > sma50 {
                score += 10.0;
            } else {
                score -= 10.0;
            }
        }

        // MACD
        if let (Some(macd), Some(signal)) = (tech.macd, tech.macd_signal) {
            if macd > signal {
                score += 10.0;
            } else {
                score -= 10.0;
            }
        }

        // Bollinger
        if let (Some(upper), Some(lower)) = (tech.bollinger_upper, tech.bollinger_lower) {
            if price < lower {
                score += 15.0;
            } else if price > upper {
                score -= 15.0;
            }
        }

        // Volume
        if let Some(vol_sma) = tech.volume_sma_20 {
            if let Some(&last_vol) = volumes.last() {
                if last_vol as f64 > vol_sma * 1.5 {
                    score += 5.0;
                }
            }
        }

        score.clamp(0.0, 100.0)
    }

    fn score_risk(&self, tech: &TechnicalIndicators, volumes: &[i64]) -> f64 {
        let mut score: f64 = 70.0;

        // Volatilité (ATR)
        if let Some(atr) = tech.atr_14 {
            if atr > 5.0 {
                score -= 20.0;
            } else if atr > 3.0 {
                score -= 10.0;
            }
        }

        // Liquidité
        if !volumes.is_empty() {
            let avg_vol = volumes[volumes.len().saturating_sub(20)..].iter().sum::<i64>() as f64
                / volumes.len().saturating_sub(20).max(1) as f64;
            if avg_vol < 1000.0 {
                score -= 15.0;
            } else if avg_vol < 5000.0 {
                score -= 5.0;
            }
        }

        score.clamp(0.0, 100.0)
    }

    fn determine_signal(
        &self,
        composite: f64,
        tech: &TechnicalIndicators,
        _fund_scores: &HashMap<String, f64>,
        reasons: &mut Vec<String>,
    ) -> (SignalType, f64) {
        let mut signal = SignalType::Hold;
        let mut confidence = 0.5;

        if composite >= 75.0 {
            if let Some(rsi) = tech.rsi_14 {
                if rsi < 40.0 {
                    signal = SignalType::StrongBuy;
                    confidence = 0.85;
                    reasons.push("Signal FORT ACHAT: Value + Survente technique".to_string());
                } else {
                    signal = SignalType::Buy;
                    confidence = 0.75;
                    reasons.push("Signal ACHAT: Fondamentaux solides".to_string());
                }
            }
        } else if composite >= 60.0 {
            signal = SignalType::Buy;
            confidence = 0.65;
            reasons.push("Signal ACHAT modéré".to_string());
        } else if composite <= 25.0 {
            if let Some(rsi) = tech.rsi_14 {
                if rsi > 65.0 {
                    signal = SignalType::StrongSell;
                    confidence = 0.80;
                    reasons.push("Signal FORTE VENTE: Faible valeur + Surachat".to_string());
                } else {
                    signal = SignalType::Sell;
                    confidence = 0.70;
                    reasons.push("Signal VENTE: Fondamentaux faibles".to_string());
                }
            }
        } else if composite <= 40.0 {
            signal = SignalType::Sell;
            confidence = 0.60;
            reasons.push("Signal VENTE modéré".to_string());
        } else {
            reasons.push("Signal NEUTRE: Attendre une meilleure opportunité".to_string());
        }

        (signal, confidence)
    }
}

// ─────────────────────────────────────────
// PARALLEL SCAN ( Rayon )
// ─────────────────────────────────────────

use rayon::prelude::*;

pub fn scan_all_tickers(
    tickers: Vec<Ticker>,
    data_map: HashMap<String, (Vec<f64>, Vec<i64>, Vec<f64>, Vec<f64>)>,
    fund_map: HashMap<String, FundamentalData>,
    current_prices: HashMap<String, f64>,
) -> Vec<ScoringResult> {
    let engine = SignalGenerationService::new();

    tickers
        .into_par_iter()
        .filter_map(|ticker| {
            let symbol = ticker.symbol.clone();
            let (closes, volumes, highs, lows) = data_map.get(&symbol)?;
            let fundamental = fund_map.get(&symbol).cloned().unwrap_or_default();
            let current_price = *current_prices.get(&symbol)?;

            Some(engine.generate_signal(
                ticker,
                closes,
                volumes,
                highs,
                lows,
                &fundamental,
                current_price,
            ))
        })
        .collect()
}

// ─────────────────────────────────────────
// TESTS
// ─────────────────────────────────────────

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
        // Prix en hausse constante → RSI proche de 100
        let prices_up: Vec<f64> = (0..20).map(|i| 100.0 + i as f64).collect();
        let rsi = service.rsi(&prices_up, 14);
        assert!(rsi > 70.0);

        // Prix en baisse constante → RSI proche de 0
        let prices_down: Vec<f64> = (0..20).map(|i| 100.0 - i as f64).collect();
        let rsi = service.rsi(&prices_down, 14);
        assert!(rsi < 30.0);
    }

    #[test]
    fn test_fundamental_scoring() {
        let service = FundamentalScoringService::new();
        let fund = FundamentalData {
            per: Some(7.5),
            roe: Some(26.0),
            dividend_yield: Some(7.5),
            debt_to_equity: Some(0.3),
            revenue_growth: Some(22.0),
            ..Default::default()
        };
        let (scores, reasons) = service.score(&fund);
        assert_eq!(scores.get("per").unwrap(), &95.0);
        assert_eq!(scores.get("roe").unwrap(), &95.0);
        assert_eq!(scores.get("dividend").unwrap(), &95.0);
        assert_eq!(scores.get("debt").unwrap(), &90.0);
        assert_eq!(scores.get("growth").unwrap(), &90.0);
        assert!(!reasons.is_empty());
    }

    #[test]
    fn test_signal_generation() {
        let service = SignalGenerationService::new();
        let ticker = Ticker {
            symbol: "SNTS".to_string(),
            name: "Sonatel".to_string(),
            country: "Sénégal".to_string(),
            sector: "Télécom".to_string(),
        };

        // Générer 60 prix avec une légère tendance haussière
        let prices: Vec<f64> = (0..60).map(|i| {
            25000.0 + (i as f64 * 50.0) + (i as f64 * 10.0 * (i % 3) as f64)
        }).collect();
        let volumes: Vec<i64> = vec![1000; 60];
        let highs: Vec<f64> = prices.iter().map(|p| p + 100.0).collect();
        let lows: Vec<f64> = prices.iter().map(|p| p - 100.0).collect();

        let fund = FundamentalData {
            per: Some(12.5),
            roe: Some(28.5),
            dividend_yield: Some(6.5),
            debt_to_equity: Some(0.4),
            revenue_growth: Some(15.0),
            ..Default::default()
        };

        let result = service.generate_signal(ticker, &prices, &volumes, &highs, &lows, &fund, 26000.0);
        assert!(result.composite_score > 50.0);
        println!("Signal: {:?}, Score: {:.1}", result.signal, result.composite_score);
        println!("Reasons: {:?}", result.reasons);
    }

    #[test]
    fn test_parallel_scan() {
        let tickers = vec![
            Ticker { symbol: "SNTS".to_string(), name: "Sonatel".to_string(), country: "Sénégal".to_string(), sector: "Télécom".to_string() },
            Ticker { symbol: "SGBC".to_string(), name: "SGB".to_string(), country: "CI".to_string(), sector: "Finance".to_string() },
            Ticker { symbol: "ORAC".to_string(), name: "Orange".to_string(), country: "CI".to_string(), sector: "Télécom".to_string() },
        ];

        let mut data_map = HashMap::new();
        let mut fund_map = HashMap::new();
        let mut current_prices = HashMap::new();

        for ticker in &tickers {
            let prices: Vec<f64> = (0..60).map(|i| 10000.0 + i as f64 * 100.0).collect();
            let volumes: Vec<i64> = vec![5000; 60];
            let highs: Vec<f64> = prices.iter().map(|p| p + 50.0).collect();
            let lows: Vec<f64> = prices.iter().map(|p| p - 50.0).collect();
            data_map.insert(ticker.symbol.clone(), (prices, volumes, highs, lows));

            fund_map.insert(ticker.symbol.clone(), FundamentalData {
                per: Some(10.0),
                roe: Some(20.0),
                dividend_yield: Some(5.0),
                ..Default::default()
            });

            current_prices.insert(ticker.symbol.clone(), 15000.0);
        }

        let results = scan_all_tickers(tickers, data_map, fund_map, current_prices);
        assert_eq!(results.len(), 3);
        for r in &results {
            println!("{}: {} (Score: {:.1})", r.ticker.symbol, r.signal.to_string(), r.composite_score);
        }
    }
}
