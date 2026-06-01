// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

use std::collections::HashMap;
use crate::domain::*;
use super::technical::TechnicalAnalysisService;
use super::fundamental::FundamentalScoringService;
use super::macro_scoring::MacroScoringService;
use super::diversification::DiversificationService;
use super::seasonality::{SeasonalityService, current_month};
use super::position_sizing::PositionSizingService;

/// The top-level signal generation pipeline for BRVM.
///
/// Scoring weights (sum = 1.0):
/// ┌─────────────────────┬────────┬──────────────────────────────────────────┐
/// │ Component           │ Weight │ Rationale                                │
/// ├─────────────────────┼────────┼──────────────────────────────────────────┤
/// │ Technical           │  20%   │ Trend/momentum; reduced (illiquid mkt)  │
/// │ Fundamental         │  25%   │ PER, ROE, Div Yield, Growth, Debt       │
/// │ Macro               │  20%   │ BCEAO, FX, commodities, political risk  │
/// │ Risk / Liquidity    │  15%   │ ATR, Amihud, Zero-Return Ratio          │
/// │ Seasonality         │  10%   │ UEMOA harvest calendar + ex-div effect  │
/// │ Diversification     │  10%   │ Portfolio exposure constraints           │
/// └─────────────────────┴────────┴──────────────────────────────────────────┘
pub struct SignalGenerationService {
    tech:        TechnicalAnalysisService,
    fund:        FundamentalScoringService,
    macro_svc:   MacroScoringService,
    div_svc:     DiversificationService,
    season_svc:  SeasonalityService,
    sizing_svc:  PositionSizingService,
}

impl Default for SignalGenerationService {
    fn default() -> Self {
        Self::new()
    }
}

impl SignalGenerationService {
    pub fn new() -> Self {
        Self {
            tech:       TechnicalAnalysisService::new(),
            fund:       FundamentalScoringService::new(),
            macro_svc:  MacroScoringService::new(),
            div_svc:    DiversificationService::new(),
            season_svc: SeasonalityService::new(),
            sizing_svc: PositionSizingService::new(),
        }
    }

    /// Convenience wrapper — no macro data, no portfolio.
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
        self.generate_signal_with_macro(
            ticker, closes, volumes, highs, lows,
            fundamental, current_price, &MacroData::default(), &[],
        )
    }

    /// Full signal generation pipeline including macro context and portfolio constraints.
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
        // ── Step 1: Compute all indicators ───────────────────────────────────
        let tech_indicators = self.tech.calculate(closes, volumes, highs, lows);
        let (fund_scores, mut reasons) = self.fund.score(fundamental);

        // ── Step 2: Score each dimension ─────────────────────────────────────
        let tech_score = self.score_technical(&tech_indicators, current_price, closes, volumes);
        let risk_score = self.score_risk(&tech_indicators, volumes);
        let (macro_score, macro_reasons) =
            self.macro_svc.score(macro_data, &ticker.country, &ticker.sector, &ticker.symbol);

        let month = current_month();
        let (seasonal_score, seasonal_reasons) = self.season_svc.score(
            &ticker.symbol, &ticker.sector, month, fundamental.dividend_yield,
        );

        // ── Step 3: Fundamental composite ────────────────────────────────────
        let fund_composite =
            fund_scores.get("per").unwrap_or(&50.0)      * 0.25 +
            fund_scores.get("roe").unwrap_or(&50.0)      * 0.25 +
            fund_scores.get("dividend").unwrap_or(&50.0) * 0.20 +
            fund_scores.get("growth").unwrap_or(&50.0)   * 0.15 +
            fund_scores.get("debt").unwrap_or(&50.0)     * 0.15;

        // ── Step 4: Preliminary composite (placeholder 50 for diversification)
        let preliminary_composite =
            tech_score     * 0.20 +
            fund_composite * 0.25 +
            macro_score    * 0.20 +
            risk_score     * 0.15 +
            seasonal_score * 0.10 +
            50.0           * 0.10;

        // ── Step 5: Diversification adjustment ───────────────────────────────
        let (diversification_score, div_reasons) =
            self.div_svc.adjust_for_diversification(preliminary_composite, &ticker, user_portfolio);

        // ── Step 6: Final composite with real diversification score ───────────
        let final_composite =
            tech_score           * 0.20 +
            fund_composite       * 0.25 +
            macro_score          * 0.20 +
            risk_score           * 0.15 +
            seasonal_score       * 0.10 +
            diversification_score * 0.10;

        // ── Step 7: Aggregate reasons ─────────────────────────────────────────
        reasons.extend(macro_reasons.clone());
        reasons.extend(div_reasons);
        reasons.extend(seasonal_reasons);

        // ── Step 8: Determine signal + confidence ─────────────────────────────
        let (signal, confidence) = self.determine_signal(
            final_composite, &tech_indicators, &fund_scores, macro_data, &mut reasons,
        );

        // ── Step 9: Kelly position sizing (BUY signals only) ─────────────────
        let (kelly_fraction, suggested_position_pct) = match signal {
            SignalType::Buy | SignalType::StrongBuy => self.sizing_svc.compute(
                &signal,
                tech_indicators.atr_14,
                current_price,
                tech_indicators.amihud_20,
                tech_indicators.zero_return_ratio,
            ),
            _ => (0.0, 0.0),
        };

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
            seasonality_score: seasonal_score,
            kelly_fraction,
            suggested_position_pct,
        }
    }

    // ─── Private scoring helpers ──────────────────────────────────────────────

    fn score_technical(
        &self,
        tech: &TechnicalIndicators,
        price: f64,
        _prices: &[f64],
        volumes: &[i64],
    ) -> f64 {
        let mut score: f64 = 50.0;

        // RSI: oversold → buy signal, overbought → sell signal
        if let Some(rsi) = tech.rsi_14 {
            if rsi < 30.0 {
                score += 20.0; // strong oversold
            } else if rsi < 40.0 {
                score += 10.0; // mild oversold
            } else if rsi > 70.0 {
                score -= 20.0; // strong overbought
            } else if rsi > 60.0 {
                score -= 8.0;  // mild overbought
            } else if (40.0..=60.0).contains(&rsi) {
                score += 5.0;  // neutral zone — slight momentum neutral
            }
        }

        // SMA crossover (Golden Cross / Death Cross)
        if let (Some(sma20), Some(sma50)) = (tech.sma_20, tech.sma_50) {
            if sma20 > sma50 {
                score += 10.0; // golden cross — uptrend
            } else {
                score -= 10.0; // death cross — downtrend
            }
        }

        // MACD crossover
        if let (Some(macd), Some(signal)) = (tech.macd, tech.macd_signal) {
            if macd > signal {
                score += 10.0; // bullish momentum
            } else {
                score -= 10.0; // bearish momentum
            }
        }

        // Bollinger Bands: mean-reversion signal
        if let (Some(upper), Some(lower)) = (tech.bollinger_upper, tech.bollinger_lower) {
            if price < lower {
                score += 15.0; // below lower band: statistically cheap
            } else if price > upper {
                score -= 15.0; // above upper band: statistically expensive
            }
        }

        // Volume confirmation: high volume validates the signal
        if let Some(vol_sma) = tech.volume_sma_20 {
            if let Some(&last_vol) = volumes.last() {
                if last_vol as f64 > vol_sma * 2.0 {
                    score += 8.0; // strong volume spike — institutional interest
                } else if last_vol as f64 > vol_sma * 1.5 {
                    score += 5.0; // moderate volume above average
                }
            }
        }

        score.clamp(0.0, 100.0)
    }

    /// Liquidity and volatility risk score.
    ///
    /// Replaces the naive volume threshold with the Amihud ratio + zero-return ratio.
    /// ATR remains as a raw volatility measure.
    fn score_risk(&self, tech: &TechnicalIndicators, volumes: &[i64]) -> f64 {
        let mut score: f64 = 70.0;

        // ATR: high volatility → higher risk
        if let Some(atr) = tech.atr_14 {
            if atr > 5.0 {
                score -= 20.0;
            } else if atr > 3.0 {
                score -= 10.0;
            }
        }

        // Amihud illiquidity: primary liquidity filter for BRVM
        if let Some(amihud) = tech.amihud_20 {
            if amihud > 2.0 {
                score -= 30.0; // execution would move the market significantly
            } else if amihud > 1.0 {
                score -= 20.0;
            } else if amihud > 0.5 {
                score -= 10.0;
            }
            // amihud < 0.1 → very liquid, slight bonus
            if amihud < 0.1 {
                score += 5.0;
            }
        } else {
            // Fallback to raw volume if Amihud data not yet available
            if !volumes.is_empty() {
                let avg_vol = volumes[volumes.len().saturating_sub(20)..]
                    .iter().sum::<i64>() as f64
                    / volumes.len().saturating_sub(20).max(1) as f64;
                if avg_vol < 1000.0 {
                    score -= 15.0;
                } else if avg_vol < 5000.0 {
                    score -= 5.0;
                }
            }
        }

        // Zero-return ratio: stale price → liquidity trap risk
        if let Some(zrr) = tech.zero_return_ratio {
            if zrr > 0.60 {
                score -= 25.0; // market effectively frozen for this ticker
            } else if zrr > 0.40 {
                score -= 15.0;
            } else if zrr > 0.20 {
                score -= 5.0;
            }
        }

        score.clamp(0.0, 100.0)
    }

    /// Translates composite score + technical context into a trading signal.
    ///
    /// Includes hard safety overrides:
    /// 1. Political crisis → force Hold regardless of score
    /// 2. Zero-return ratio > 0.50 → force Hold (market frozen)
    fn determine_signal(
        &self,
        composite: f64,
        tech: &TechnicalIndicators,
        _fund_scores: &HashMap<String, f64>,
        macro_data: &MacroData,
        reasons: &mut Vec<String>,
    ) -> (SignalType, f64) {

        // ── Hard override #1: Political crisis ────────────────────────────────
        if macro_data.political_crisis {
            reasons.push("HOLD forcé: crise politique active dans le pays d'origine".to_string());
            return (SignalType::Hold, 0.30);
        }

        // ── Hard override #2: Frozen / illiquid market ────────────────────────
        if let Some(zrr) = tech.zero_return_ratio {
            if zrr > 0.50 {
                reasons.push(format!(
                    "HOLD forcé: {:.0}% de jours sans variation de prix — marché illiquide",
                    zrr * 100.0
                ));
                return (SignalType::Hold, 0.25);
            }
        }

        // ── Hard override #3: Extremely illiquid (Amihud) ─────────────────────
        if let Some(amihud) = tech.amihud_20 {
            if amihud > 5.0 {
                reasons.push(format!(
                    "HOLD forcé: ratio Amihud {:.2} — impact de marché trop élevé pour exécuter",
                    amihud
                ));
                return (SignalType::Hold, 0.20);
            }
        }

        // ── Standard signal determination based on composite score ─────────────
        match composite {
            s if s >= 75.0 => {
                // Strong buy zone — confirm with RSI to avoid buying overbought stocks
                if let Some(rsi) = tech.rsi_14 {
                    if rsi < 40.0 {
                        reasons.push("✅ FORT ACHAT: Score élevé + survente technique confirmée".to_string());
                        (SignalType::StrongBuy, 0.85)
                    } else if rsi > 70.0 {
                        // High score but overbought — downgrade to plain Buy
                        reasons.push("⚠️ ACHAT modéré: bon score mais RSI en surachat".to_string());
                        (SignalType::Buy, 0.60)
                    } else {
                        reasons.push("✅ ACHAT: Fondamentaux solides, contexte macro favorable".to_string());
                        (SignalType::Buy, 0.75)
                    }
                } else {
                    reasons.push("✅ ACHAT: Bon score composite (pas de données RSI)".to_string());
                    (SignalType::Buy, 0.65)
                }
            }
            s if s >= 62.0 => {
                reasons.push("📈 ACHAT modéré: score favorable — surveiller la liquidité".to_string());
                (SignalType::Buy, 0.62)
            }
            s if s <= 25.0 => {
                if let Some(rsi) = tech.rsi_14 {
                    if rsi > 65.0 {
                        reasons.push("🔴 FORTE VENTE: Score faible + surachat technique".to_string());
                        (SignalType::StrongSell, 0.82)
                    } else {
                        reasons.push("🔴 VENTE: Fondamentaux faibles ou risque macro élevé".to_string());
                        (SignalType::Sell, 0.70)
                    }
                } else {
                    reasons.push("🔴 VENTE: Score composite insuffisant".to_string());
                    (SignalType::Sell, 0.60)
                }
            }
            s if s <= 40.0 => {
                reasons.push("📉 VENTE modérée: score défavorable".to_string());
                (SignalType::Sell, 0.58)
            }
            _ => {
                reasons.push("⏸ NEUTRE: Attendre une meilleure opportunité".to_string());
                (SignalType::Hold, 0.50)
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn make_trending_prices(n: usize, start: f64, step: f64) -> Vec<f64> {
        (0..n).map(|i| start + i as f64 * step).collect()
    }

    #[test]
    fn test_signal_generation_bullish() {
        let service = SignalGenerationService::new();
        let ticker = Ticker {
            symbol: "SNTS".to_string(),
            name: "Sonatel".to_string(),
            country: "Sénégal".to_string(),
            sector: "Télécom".to_string(),
        };
        let prices = make_trending_prices(60, 25000.0, 50.0);
        let volumes: Vec<i64> = vec![10000; 60];
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

        let result = service.generate_signal(ticker, &prices, &volumes, &highs, &lows, &fund, 28000.0);
        assert!(result.composite_score > 50.0, "Expected bullish score, got {:.1}", result.composite_score);
        println!("Signal: {}, Score: {:.1}, Kelly: {:.1}%", result.signal, result.composite_score, result.suggested_position_pct);
        println!("Reasons: {:?}", result.reasons);
    }

    #[test]
    fn test_political_crisis_forces_hold() {
        let service = SignalGenerationService::new();
        let ticker = Ticker {
            symbol: "ONTBF".to_string(),
            name: "Onatel BF".to_string(),
            country: "Burkina Faso".to_string(),
            sector: "Télécom".to_string(),
        };
        let prices = make_trending_prices(60, 10000.0, 100.0); // strongly bullish
        let volumes: Vec<i64> = vec![5000; 60];
        let highs: Vec<f64> = prices.iter().map(|p| p + 200.0).collect();
        let lows: Vec<f64> = prices.iter().map(|p| p - 200.0).collect();

        let fund = FundamentalData { per: Some(5.0), roe: Some(30.0), dividend_yield: Some(8.0), ..Default::default() };
        let macro_data = MacroData { political_crisis: true, ..Default::default() };

        let result = service.generate_signal_with_macro(
            ticker, &prices, &volumes, &highs, &lows, &fund, 15000.0, &macro_data, &[],
        );
        assert!(matches!(result.signal, SignalType::Hold), "Political crisis must force Hold");
        assert_eq!(result.kelly_fraction, 0.0);
        println!("Correctly blocked: {:?}", result.reasons);
    }

    #[test]
    fn test_frozen_market_forces_hold() {
        let service = SignalGenerationService::new();
        let ticker = Ticker {
            symbol: "SDSC".to_string(),
            name: "SDSC".to_string(),
            country: "CI".to_string(),
            sector: "Industrie".to_string(),
        };
        // Flat prices — zero_return_ratio will be ~1.0
        let prices = vec![10000.0_f64; 60];
        let volumes: Vec<i64> = vec![100; 60];
        let highs = prices.clone();
        let lows = prices.clone();
        let fund = FundamentalData { per: Some(6.0), roe: Some(25.0), ..Default::default() };

        let result = service.generate_signal(ticker, &prices, &volumes, &highs, &lows, &fund, 10000.0);
        assert!(matches!(result.signal, SignalType::Hold), "Frozen market must force Hold, got: {}", result.signal);
        assert_eq!(result.kelly_fraction, 0.0);
    }

    #[test]
    fn test_kelly_positive_for_buy() {
        let service = SignalGenerationService::new();
        let ticker = Ticker {
            symbol: "SGBC".to_string(),
            name: "SGB".to_string(),
            country: "CI".to_string(),
            sector: "Finance".to_string(),
        };
        let prices = make_trending_prices(60, 15000.0, 80.0);
        let volumes: Vec<i64> = vec![20000; 60];
        let highs: Vec<f64> = prices.iter().map(|p| p + 150.0).collect();
        let lows: Vec<f64> = prices.iter().map(|p| p - 150.0).collect();

        let fund = FundamentalData {
            per: Some(8.0), roe: Some(22.0), dividend_yield: Some(5.0),
            debt_to_equity: Some(0.3), revenue_growth: Some(12.0), ..Default::default()
        };

        let result = service.generate_signal(ticker, &prices, &volumes, &highs, &lows, &fund, 19800.0);
        if matches!(result.signal, SignalType::Buy | SignalType::StrongBuy) {
            assert!(result.kelly_fraction > 0.0, "Buy signal must have positive Kelly fraction");
            assert!(result.suggested_position_pct <= 25.0);
            println!("Kelly: {:.1}%, Position: {:.1}%", result.kelly_fraction * 100.0, result.suggested_position_pct);
        }
    }
}
