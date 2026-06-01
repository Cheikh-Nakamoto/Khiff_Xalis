// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

use std::collections::HashMap;
use crate::domain::*;
use super::technical::TechnicalAnalysisService;
use super::fundamental::FundamentalScoringService;
use super::macro_scoring::MacroScoringService;
use super::diversification::DiversificationService;

pub struct SignalGenerationService {
    tech: TechnicalAnalysisService,
    fund: FundamentalScoringService,
    macro_svc: MacroScoringService,
    div_svc: DiversificationService,
}

impl Default for SignalGenerationService {
    fn default() -> Self {
        Self::new()
    }
}

impl SignalGenerationService {
    pub fn new() -> Self {
        Self {
            tech: TechnicalAnalysisService::new(),
            fund: FundamentalScoringService::new(),
            macro_svc: MacroScoringService::new(),
            div_svc: DiversificationService::new(),
        }
    }

    #[allow(clippy::too_many_arguments)]
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
        self.generate_signal_with_macro(ticker, closes, volumes, highs, lows, fundamental, current_price, &MacroData::default(), &[])
    }

    #[allow(clippy::too_many_arguments)]
    pub fn generate_signal_with_macro(
        &self,
        ticker: Ticker,
        closes: &[f64],
        volumes: &[i64],
        highs: &[f64],
        lows: &[f64],
        fundamental: &FundamentalData,
        current_price: f64,
        macro_data: &MacroData,
        user_portfolio: &[PortfolioPosition],
    ) -> ScoringResult {
        let tech_indicators = self.tech.calculate(closes, volumes, highs, lows);
        let (fund_scores, mut reasons) = self.fund.score(fundamental);
        let tech_score = self.score_technical(&tech_indicators, current_price, closes, volumes);
        let risk_score = self.score_risk(&tech_indicators, volumes);
        let (macro_score, macro_reasons) = self.macro_svc.score(macro_data, &ticker.country, &ticker.sector);

        let fund_composite =
            fund_scores.get("per").unwrap_or(&50.0) * 0.25 +
            fund_scores.get("roe").unwrap_or(&50.0) * 0.25 +
            fund_scores.get("dividend").unwrap_or(&50.0) * 0.20 +
            fund_scores.get("growth").unwrap_or(&50.0) * 0.15 +
            fund_scores.get("debt").unwrap_or(&50.0) * 0.15;

        // Weights: tech=0.25, fund=0.25, macro=0.20, risk=0.15, div=0.15 → sum=1.0
        let preliminary_composite =
            tech_score * 0.25 +
            fund_composite * 0.25 +
            macro_score * 0.20 +
            risk_score * 0.15 +
            50.0 * 0.15; // placeholder for diversification before it's computed

        let (diversification_score, div_reasons) = self.div_svc.adjust_for_diversification(preliminary_composite, &ticker, user_portfolio);

        let final_composite =
            tech_score * 0.25 +
            fund_composite * 0.25 +
            macro_score * 0.20 +
            risk_score * 0.15 +
            diversification_score * 0.15;

        reasons.extend(macro_reasons.clone());
        reasons.extend(div_reasons);

        let (signal, confidence) = self.determine_signal(final_composite, &tech_indicators, &fund_scores, &mut reasons);

        ScoringResult {
            ticker,
            composite_score: final_composite,
            signal,
            confidence,
            reasons,
            technical_indicators: tech_indicators,
            fundamental_scores: fund_scores,
            technical_score: tech_score,
            risk_score,
            macro_score,
            diversification_score,
            macro_reasons,
        }
    }

    fn score_technical(&self, tech: &TechnicalIndicators, price: f64, _prices: &[f64], volumes: &[i64]) -> f64 {
        let mut score: f64 = 50.0;

        if let Some(rsi) = tech.rsi_14 {
            if rsi < 30.0 {
                score += 20.0;
            } else if rsi > 70.0 {
                score -= 20.0;
            } else if (40.0..=60.0).contains(&rsi) {
                score += 5.0;
            }
        }

        if let (Some(sma20), Some(sma50)) = (tech.sma_20, tech.sma_50) {
            if sma20 > sma50 {
                score += 10.0;
            } else {
                score -= 10.0;
            }
        }

        if let (Some(macd), Some(signal)) = (tech.macd, tech.macd_signal) {
            if macd > signal {
                score += 10.0;
            } else {
                score -= 10.0;
            }
        }

        if let (Some(upper), Some(lower)) = (tech.bollinger_upper, tech.bollinger_lower) {
            if price < lower {
                score += 15.0;
            } else if price > upper {
                score -= 15.0;
            }
        }

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

        if let Some(atr) = tech.atr_14 {
            if atr > 5.0 {
                score -= 20.0;
            } else if atr > 3.0 {
                score -= 10.0;
            }
        }

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
            } else {
                signal = SignalType::Buy;
                confidence = 0.65;
                reasons.push("Signal ACHAT: Bon score composite mais manque de données RSI".to_string());
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
            } else {
                signal = SignalType::Sell;
                confidence = 0.60;
                reasons.push("Signal VENTE: Faible score composite mais manque de données RSI".to_string());
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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_signal_generation() {
        let service = SignalGenerationService::new();
        let ticker = Ticker {
            symbol: "SNTS".to_string(),
            name: "Sonatel".to_string(),
            country: "Sénégal".to_string(),
            sector: "Télécom".to_string(),
        };

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
}
