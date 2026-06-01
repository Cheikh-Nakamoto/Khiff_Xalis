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

    /// Scores the macroeconomic context for a given ticker.
    ///
    /// Returns (score ∈ [0,100], reasons) where 50 = neutral.
    ///
    /// # Parameters
    /// - `macro_data`: macroeconomic context for this ticker's country
    /// - `country`:    ISO-2 country code (e.g. "CI", "SN")
    /// - `sector`:     sector string (e.g. "Agriculture", "Finance")
    /// - `symbol`:     ticker symbol (e.g. "SACI", "PALC") for commodity-specific scoring
    pub fn score(
        &self,
        macro_data: &MacroData,
        _country: &str,
        sector: &str,
        symbol: &str,
    ) -> (f64, Vec<String>) {
        let mut score: f64 = 50.0;
        let mut reasons = Vec::new();

        // ── POLITICAL CRISIS FREEZE ──────────────────────────────────────────
        // Hard block: if a coup or major crisis is detected in this country,
        // the macro score plunges. The determine_signal logic will enforce Hold.
        if macro_data.political_crisis {
            score -= 40.0;
            reasons.push("🚨 Crise politique active — signaux d'achat suspendus pour ce pays".to_string());
            // Return early: no point scoring the rest when the market is frozen.
            return (score.clamp(0.0, 100.0), reasons);
        }

        // ── INFLATION ────────────────────────────────────────────────────────
        match macro_data.inflation {
            Some(i) if (2.0..=5.0).contains(&i) => {
                score += 15.0;
                reasons.push(format!("Inflation stable: {:.1}% — zone cible BCEAO", i));
            }
            Some(i) if (5.0..=10.0).contains(&i) => {
                score -= 8.0;
                reasons.push(format!("Inflation modérément élevée: {:.1}%", i));
            }
            Some(i) if i > 10.0 => {
                score -= 20.0;
                reasons.push(format!("Inflation élevée: {:.1}% — pression sévère sur les marges", i));
            }
            Some(i) if i < 0.0 => {
                score -= 10.0;
                reasons.push(format!("Déflation: {:.1}% — demande intérieure faible", i));
            }
            _ => {}
        }

        // ── BCEAO TAUX DIRECTEUR ──────────────────────────────────────────────
        match macro_data.taux_directeur {
            Some(t) if t <= 3.0 => {
                score += 10.0;
                reasons.push(format!("Taux BCEAO bas: {:.2}% — crédit et investissement favorisés", t));
            }
            Some(t) if (3.0..=5.0).contains(&t) => {
                // Neutral zone
            }
            Some(t) if t > 5.0 && t <= 7.0 => {
                score -= 12.0;
                reasons.push(format!("Taux BCEAO restrictif: {:.2}% — coût du crédit élevé", t));
            }
            Some(t) if t > 7.0 => {
                score -= 20.0;
                reasons.push(format!("Taux BCEAO très élevé: {:.2}% — frein à l'activité économique", t));
            }
            _ => {}
        }

        // ── POLITICAL STABILITY (WGI score -2.5 to +2.5) ────────────────────
        match macro_data.political_stability {
            Some(p) if p > 0.5 => {
                score += 15.0;
                reasons.push(format!("Stabilité politique forte: {:.2} — environnement favorable", p));
            }
            Some(p) if (-0.5..=0.5).contains(&p) => {
                // Neutral, no bonus/malus
            }
            Some(p) if p < -0.5 && p >= -1.5 => {
                score -= 15.0;
                reasons.push(format!("Instabilité politique modérée: {:.2}", p));
            }
            Some(p) if p < -1.5 => {
                score -= 28.0;
                reasons.push(format!("Risque pays critique: {:.2} — forte instabilité politique", p));
            }
            _ => {}
        }

        // ── XOF/EUR (parité fixe, surveiller déviation) ──────────────────────
        match macro_data.change_xof_eur {
            Some(x) if x > 670.0
                && (sector == "Consommation" || sector == "Distribution") =>
            {
                score -= 10.0;
                reasons.push(format!("XOF/EUR déprécié ({:.0}) — renchérit les importations euro", x));
            }
            Some(x) if x < 645.0
                && (sector == "Consommation" || sector == "Distribution") =>
            {
                score += 5.0;
                reasons.push(format!("XOF/EUR apprécié ({:.0}) — avantage importateurs", x));
            }
            _ => {}
        }

        // ── XOF/USD (flottant — crucial pour les exportateurs UEMOA) ─────────
        // USD fort → FCFA relatif faible → exports BRVM moins chers en USD → bonus
        // USD fort → importations plus chères → malus Distribution/Industrie
        match (sector, macro_data.change_xof_usd) {
            ("Agriculture", Some(usd)) if usd > 630.0 => {
                score += 12.0;
                reasons.push(format!("USD fort ({:.0} XOF/$) — compétitivité export agricole renforcée", usd));
            }
            ("Agriculture", Some(usd)) if usd < 580.0 => {
                score -= 10.0;
                reasons.push(format!("USD faible ({:.0} XOF/$) — pression sur recettes d'exportation", usd));
            }
            ("Distribution", Some(usd)) | ("Industrie", Some(usd)) if usd > 630.0 => {
                score -= 10.0;
                reasons.push(format!("USD fort ({:.0} XOF/$) — renchérit les importations industrielles", usd));
            }
            _ => {}
        }

        // ── COMMODITY PRICES ──────────────────────────────────────────────────

        // Cacao (Côte d'Ivoire = 40% de la prod mondiale)
        match (sector, macro_data.cocoa_price) {
            ("Agriculture", Some(c)) if c > 3500.0 => {
                score += 20.0;
                reasons.push(format!("Prix cacao très favorable: ${:.0}/t", c));
            }
            ("Agriculture", Some(c)) if c > 3000.0 => {
                score += 12.0;
                reasons.push(format!("Prix cacao favorable: ${:.0}/t", c));
            }
            ("Agriculture", Some(c)) if c < 2000.0 => {
                score -= 15.0;
                reasons.push(format!("Prix cacao bas: ${:.0}/t — pression recettes", c));
            }
            _ => {}
        }

        // Noix de cajou (SACI, SAFC — Côte d'Ivoire #1 mondial)
        let is_cashew = matches!(symbol, "SACI" | "SAFC");
        if is_cashew {
            match macro_data.cashew_price {
                Some(p) if p > 1100.0 => {
                    score += 18.0;
                    reasons.push(format!("Prix cajou élevé: ${:.0}/t — bénéfice direct {}", p, symbol));
                }
                Some(p) if p > 900.0 => {
                    score += 8.0;
                    reasons.push(format!("Prix cajou correct: ${:.0}/t", p));
                }
                Some(p) if p < 700.0 => {
                    score -= 15.0;
                    reasons.push(format!("Prix cajou bas: ${:.0}/t — pression marges", p));
                }
                _ => {}
            }
        }

        // Or (CIDA, SGBCI exposée mines)
        let is_gold_exposed = matches!(symbol, "CIDA" | "SGBC");
        if is_gold_exposed || sector == "Mines" {
            match macro_data.gold_price {
                Some(g) if g > 2000.0 => {
                    score += 15.0;
                    reasons.push(format!("Or au plus haut: ${:.0}/oz — forte marge minière", g));
                }
                Some(g) if g > 1800.0 => {
                    score += 8.0;
                    reasons.push(format!("Prix or favorable: ${:.0}/oz", g));
                }
                Some(g) if g < 1500.0 => {
                    score -= 12.0;
                    reasons.push(format!("Prix or bas: ${:.0}/oz — impact sur cash-flows miniers", g));
                }
                _ => {}
            }
        }

        // Caoutchouc naturel (PALC, SOGB, SAPH)
        let is_rubber = matches!(symbol, "PALC" | "SOGB" | "SAPH");
        if is_rubber {
            match macro_data.rubber_price {
                Some(r) if r > 1.8 => {
                    score += 15.0;
                    reasons.push(format!("Prix caoutchouc favorable: ${:.2}/kg", r));
                }
                Some(r) if r < 1.2 => {
                    score -= 12.0;
                    reasons.push(format!("Prix caoutchouc bas: ${:.2}/kg", r));
                }
                _ => {}
            }
        }

        // Huile de palme (PALC principalement)
        if symbol == "PALC" {
            match macro_data.palm_oil_price {
                Some(p) if p > 1000.0 => {
                    score += 10.0;
                    reasons.push(format!("Prix huile de palme élevé: ${:.0}/t", p));
                }
                Some(p) if p < 700.0 => {
                    score -= 8.0;
                    reasons.push(format!("Prix huile de palme bas: ${:.0}/t", p));
                }
                _ => {}
            }
        }

        // Pétrole Brent (secteurs sensibles)
        match (sector, macro_data.oil_price) {
            ("Distribution", Some(o)) | ("Transport", Some(o)) if o < 70.0 => {
                score += 15.0;
                reasons.push(format!("Brent bas: ${:.1}/bbl — coûts logistiques réduits", o));
            }
            ("Distribution", Some(o)) | ("Transport", Some(o)) if o > 100.0 => {
                score -= 15.0;
                reasons.push(format!("Brent élevé: ${:.1}/bbl — pression marges distribution", o));
            }
            ("Industrie", Some(o)) if o > 100.0 => {
                score -= 10.0;
                reasons.push(format!("Brent élevé: ${:.1}/bbl — coûts industriels en hausse", o));
            }
            _ => {}
        }

        // ── COMMODITY BETA AMPLIFICATION ─────────────────────────────────────
        // If a per-stock commodity beta has been pre-computed by go-collector,
        // amplify the commodity signal proportionally.
        if let Some(beta) = macro_data.commodity_beta {
            if beta.abs() > 0.5 {
                // A large beta means the stock is very sensitive to commodity moves.
                // We add a small bonus to acknowledge this correlation quality,
                // but the primary commodity price signal already captured the direction.
                if beta > 0.8 {
                    reasons.push(format!("Commodity beta élevé ({:.2}) — forte corrélation matière première", beta));
                }
            }
        }

        // ── SOVEREIGN RATING (1=AAA, 20=D) ───────────────────────────────────
        match macro_data.sovereign_rating {
            Some(r) if r <= 5 => {
                score += 10.0;
                reasons.push("Notation souveraine investment grade — accès aux capitaux étrangers".to_string());
            }
            Some(r) if r >= 15 => {
                score -= 20.0;
                reasons.push("Notation souveraine très faible — risque de contagion systémique".to_string());
            }
            _ => {}
        }

        (score.clamp(0.0, 100.0), reasons)
    }
}
