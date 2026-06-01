// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

use crate::domain::MacroData;

pub struct MacroScoringService;

impl Default for MacroScoringService {
    fn default() -> Self {
        Self::new()
    }
}

impl MacroScoringService {
    pub fn new() -> Self {
        Self
    }

    pub fn score(&self, macro_data: &MacroData, _country: &str, sector: &str) -> (f64, Vec<String>) {
        let mut score: f64 = 50.0;
        let mut reasons = Vec::new();

        // Inflation modérée (2-5%) favorable
        match macro_data.inflation {
            Some(i) if (2.0..=5.0).contains(&i) => {
                score += 15.0;
                reasons.push(format!("Inflation stable: {:.1}%", i));
            }
            Some(i) if i > 10.0 => {
                score -= 20.0;
                reasons.push(format!("Inflation élevée: {:.1}% — pression sur marges", i));
            }
            Some(i) if i < 0.0 => {
                score -= 10.0;
                reasons.push(format!("Déflation: {:.1}% — demande faible", i));
            }
            _ => {}
        }

        // Taux directeur BCEAO
        match macro_data.taux_directeur {
            Some(t) if t <= 3.0 => {
                score += 10.0;
                reasons.push(format!("Taux directeur bas: {:.1}% — crédit favorable", t));
            }
            Some(t) if t > 6.0 => {
                score -= 15.0;
                reasons.push(format!("Taux directeur élevé: {:.1}% — coût du crédit", t));
            }
            _ => {}
        }

        // Stabilité politique (WGI: -2.5 à +2.5)
        match macro_data.political_stability {
            Some(p) if p > 0.5 => {
                score += 15.0;
                reasons.push(format!("Stabilité politique forte: {:.2}", p));
            }
            Some(p) if p < -0.5 => {
                score -= 25.0;
                reasons.push(format!("Instabilité politique: {:.2} — risque pays", p));
            }
            _ => {}
        }

        // Prix matières premières sectorielles
        match (sector, macro_data.cocoa_price, macro_data.oil_price) {
            ("Agriculture", Some(c), _) if c > 3000.0 => {
                score += 20.0;
                reasons.push(format!("Prix cacao favorable: ${:.0}/t", c));
            }
            ("Agriculture", Some(c), _) if c < 2000.0 => {
                score -= 15.0;
                reasons.push(format!("Prix cacao bas: ${:.0}/t", c));
            }
            ("Distribution", _, Some(o)) if o < 70.0 => {
                score += 15.0;
                reasons.push(format!("Prix pétrole bas: ${:.1} — coûts réduits", o));
            }
            ("Distribution", _, Some(o)) if o > 100.0 => {
                score -= 15.0;
                reasons.push(format!("Prix pétrole élevé: ${:.1} — pression marges", o));
            }
            ("Industrie", _, Some(o)) if o > 100.0 => {
                score -= 10.0;
                reasons.push(format!("Prix pétrole élevé: ${:.1} — coûts industriels", o));
            }
            _ => {}
        }

        // Notation souveraine (1=AAA, 20=D)
        match macro_data.sovereign_rating {
            Some(r) if r <= 5 => {
                score += 10.0;
                reasons.push("Notation souveraine solide".to_string());
            }
            Some(r) if r >= 15 => {
                score -= 20.0;
                reasons.push("Notation souveraine faible — risque systémique".to_string());
            }
            _ => {}
        }

        // Taux de change XOF/EUR (655.957 = parité fixe)
        match macro_data.change_xof_eur {
            Some(x)
                if x > 670.0
                    && (sector == "Consommation" || sector == "Distribution") =>
            {
                score -= 10.0;
                reasons.push("XOF déprécié — pression importateurs".to_string());
            }
            Some(x)
                if x < 650.0
                    && (sector == "Consommation" || sector == "Distribution") =>
            {
                score += 5.0;
                reasons.push("XOF apprécié — avantage importateurs".to_string());
            }
            _ => {}
        }

        (score.clamp(0.0, 100.0), reasons)
    }
}
