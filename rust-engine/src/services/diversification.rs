// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

use crate::domain::{Ticker, PortfolioPosition};

pub struct DiversificationService;

impl Default for DiversificationService {
    fn default() -> Self {
        Self::new()
    }
}

impl DiversificationService {
    pub fn new() -> Self {
        Self
    }

    pub fn adjust_for_diversification(
        &self,
        base_score: f64,
        ticker: &Ticker,
        user_portfolio: &[PortfolioPosition],
    ) -> (f64, Vec<String>) {
        let mut adjusted = base_score;
        let mut reasons = Vec::new();

        let total_value: f64 = user_portfolio.iter().map(|p| p.market_value).sum();

        if total_value == 0.0 {
            return (adjusted, reasons);
        }

        let country_exposure: f64 = user_portfolio
            .iter()
            .filter(|p| p.country == ticker.country)
            .map(|p| p.market_value)
            .sum();
        let country_pct = country_exposure / total_value;

        if country_pct > 0.40 {
            adjusted -= 15.0;
            reasons.push(format!(
                "Sur-exposition {}: {:.0}% — diversifier géographiquement",
                ticker.country, country_pct * 100.0
            ));
        } else if country_pct < 0.10 && ticker.country != "CI" {
            adjusted += 5.0;
            reasons.push(format!(
                "Sous-représentation {}: opportunité géographique",
                ticker.country
            ));
        }

        let sector_exposure: f64 = user_portfolio
            .iter()
            .filter(|p| p.sector == ticker.sector)
            .map(|p| p.market_value)
            .sum();
        let sector_pct = sector_exposure / total_value;

        if sector_pct > 0.35 {
            adjusted -= 10.0;
            reasons.push(format!(
                "Sur-exposition secteur {}: {:.0}% — risque sectoriel",
                ticker.sector, sector_pct * 100.0
            ));
        }

        (adjusted.clamp(0.0, 100.0), reasons)
    }
}
