// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

/// Seasonality scoring service based on UEMOA agricultural harvest calendars
/// and the BRVM ex-dividend effect.
///
/// Returns a multiplicative factor in [0.75, 1.25]:
/// - 1.0  = no seasonal effect (neutral)
/// - >1.0 = seasonal tailwind (harvest peak, pre-ex-div rally)
/// - <1.0 = seasonal headwind (off-season, post-ex-div dip)
pub struct SeasonalityService;

impl Default for SeasonalityService {
    fn default() -> Self {
        Self::new()
    }
}

impl SeasonalityService {
    pub fn new() -> Self {
        Self
    }

    /// Returns the seasonal score (50-based, same scale as other scores).
    ///
    /// Converts the multiplicative factor to an additive score contribution:
    /// factor 1.20 → +10 pts above neutral 50 → score 60
    /// factor 0.80 → -10 pts below neutral 50 → score 40
    pub fn score(&self, symbol: &str, sector: &str, month: u32, dividend_yield: Option<f64>) -> (f64, Vec<String>) {
        let mut reasons = Vec::new();
        let factor = self.seasonal_factor(symbol, sector, month, dividend_yield, &mut reasons);

        // Map factor [0.75, 1.25] → score [37.5, 62.5] centered on 50
        let score = 50.0 + (factor - 1.0) * 100.0;
        (score.clamp(0.0, 100.0), reasons)
    }

    /// Returns the raw multiplicative factor.
    pub fn seasonal_factor(
        &self,
        symbol: &str,
        sector: &str,
        month: u32,
        dividend_yield: Option<f64>,
        reasons: &mut Vec<String>,
    ) -> f64 {
        let mut factor = 1.0_f64;

        // --- Stock-specific harvest calendar overrides ---
        let stock_f = Self::stock_harvest_factor(symbol, month);
        if (stock_f - 1.0).abs() > 1e-9 {
            factor *= stock_f;
            if stock_f > 1.0 {
                reasons.push(format!("Saisonnalité {} mois {}: pic de récolte ({:.0}%)", symbol, month, (stock_f - 1.0) * 100.0));
            } else {
                reasons.push(format!("Saisonnalité {} mois {}: hors saison ({:.0}%)", symbol, month, (stock_f - 1.0) * 100.0));
            }
            return factor.clamp(0.75, 1.25);
        }

        // --- Sector-level fallback ---
        let sector_f = Self::sector_month_factor(sector, month);
        if (sector_f - 1.0).abs() > 1e-9 {
            factor *= sector_f;
            reasons.push(format!("Saisonnalité secteur {} mois {}: ({:+.0}%)", sector, month, (sector_f - 1.0) * 100.0));
        }

        // --- Ex-dividend rally effect ---
        // BRVM companies pay dividends mostly Q1 (Jan-Apr). Stocks with high yield
        // tend to rise 5 sessions before ex-date — we approximate with a calendar bonus.
        if let Some(dy) = dividend_yield {
            if dy > 5.0 && (1..=3).contains(&month) {
                factor *= 1.15;
                reasons.push(format!("Effet pré-ex-dividende: rendement {:.1}% — rally typique Jan-Mars", dy));
            } else if dy > 3.0 && (1..=4).contains(&month) {
                factor *= 1.08;
                reasons.push(format!("Légère pression pré-ex-dividende: rendement {:.1}%", dy));
            }
        }

        factor.clamp(0.75, 1.25)
    }

    /// Per-stock harvest calendar factors (UEMOA agricultural commodities).
    ///
    /// Sources: FIRCA (Côte d'Ivoire), ONUDI, BCEAO agricultural reports.
    fn stock_harvest_factor(symbol: &str, month: u32) -> f64 {
        match (symbol, month) {
            // ── Noix de cajou (Côte d'Ivoire = #1 producteur mondial) ──────────
            // Récolte: Feb-May, pic en Mars-Avril
            ("SACI", 3) | ("SAFC", 3) => 1.20,
            ("SACI", 4) | ("SAFC", 4) => 1.18,
            ("SACI", 2) | ("SAFC", 2) | ("SACI", 5) | ("SAFC", 5) => 1.10,
            ("SACI", 6..=12) | ("SAFC", 6..=12) | ("SACI", 1) | ("SAFC", 1) => 0.92,

            // ── Cacao (grande traite Oct-Mar, petite traite Avr-Sep) ────────────
            ("SOGC", 10) | ("SOGC", 11) | ("CFDA", 10) | ("CFDA", 11) => 1.15,
            ("SOGC", 12) | ("SOGC", 1) | ("CFDA", 12) | ("CFDA", 1) => 1.10,
            ("SOGC", 2) | ("SOGC", 3) | ("CFDA", 2) | ("CFDA", 3) => 1.05,
            ("SOGC", 5..=9) | ("CFDA", 5..=9) => 0.95,

            // ── Caoutchouc naturel (production max Avr-Sep, chute Oct-Jan) ──────
            ("PALC", 4) | ("PALC", 5) | ("PALC", 6) => 1.15,
            ("PALC", 7) | ("PALC", 8) | ("PALC", 9) => 1.08,
            ("PALC", 11..=12) | ("PALC", 1) | ("PALC", 2) => 0.88,
            ("SOGB", 4) | ("SOGB", 5) | ("SOGB", 6) => 1.15,
            ("SOGB", 7) | ("SOGB", 8) | ("SOGB", 9) => 1.08,
            ("SOGB", 11..=12) | ("SOGB", 1) | ("SOGB", 2) => 0.88,
            ("SAPH", 4..=6) => 1.12,
            ("SAPH", 11..=12) | ("SAPH", 1..=2) => 0.90,

            // ── Coton (récolte Oct-Déc: Burkina, Mali, Bénin, Togo) ────────────
            ("FILC", 10) | ("FILC", 11) | ("FILC", 12) => 1.10,
            ("FILC", 1) | ("FILC", 2) => 0.92,

            // ── Huile de palme (PALC déjà couvert ci-dessus via caoutchouc) ────
            // Secteur télécom/banque — pas de saisonnalité agricole marquée
            _ => 1.0,
        }
    }

    /// Sector-level seasonal adjustments (less precise than stock-level).
    fn sector_month_factor(sector: &str, month: u32) -> f64 {
        match (sector, month) {
            // Agriculture broad: bonus récolte Q4 (cacao), Q1-Q2 (cajou)
            ("Agriculture", 3) | ("Agriculture", 4) => 1.08,
            ("Agriculture", 10) | ("Agriculture", 11) => 1.10,
            // Distribution: bonus fêtes de fin d'année
            ("Distribution", 11) | ("Distribution", 12) => 1.05,
            // Finance: bonus saison de dividendes / clôtures annuelles
            ("Finance", 1) | ("Finance", 2) => 1.05,
            _ => 1.0,
        }
    }
}

/// Gets the current UTC month (1-12) using std::time only (no chrono dependency).
pub fn current_month() -> u32 {
    use std::time::{SystemTime, UNIX_EPOCH};
    let secs = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs();
    // Approximate month from unix timestamp:
    // Days since epoch → year/month via a simple algorithm
    let days = secs / 86400;
    let (_, month, _) = days_to_ymd(days);
    month
}

/// Converts days since Unix epoch (1970-01-01) to (year, month, day).
fn days_to_ymd(mut days: u64) -> (u32, u32, u32) {
    // 400-year cycle = 146097 days
    let z = days + 719468;
    let era = z / 146097;
    let doe = z - era * 146097;
    let yoe = (doe - doe / 1460 + doe / 36524 - doe / 146096) / 365;
    let y = yoe + era * 400;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = if mp < 10 { mp + 3 } else { mp - 9 };
    let y = if m <= 2 { y + 1 } else { y };
    (y as u32, m as u32, d as u32)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_cajou_peak() {
        let svc = SeasonalityService::new();
        let mut reasons = Vec::new();
        // SACI in March (cajou peak) should have factor > 1.15
        let f = svc.seasonal_factor("SACI", "Agriculture", 3, None, &mut reasons);
        assert!(f >= 1.18, "Expected f >= 1.18, got {}", f);
        assert!(!reasons.is_empty());
    }

    #[test]
    fn test_caoutchouc_peak() {
        let svc = SeasonalityService::new();
        let mut reasons = Vec::new();
        let f = svc.seasonal_factor("PALC", "Agriculture", 5, None, &mut reasons);
        assert!(f >= 1.10, "Expected f >= 1.10 for PALC in May, got {}", f);
    }

    #[test]
    fn test_offseason_penalty() {
        let svc = SeasonalityService::new();
        let mut reasons = Vec::new();
        // SACI in November (off-season for cajou) → factor < 1.0
        let f = svc.seasonal_factor("SACI", "Agriculture", 11, None, &mut reasons);
        assert!(f < 1.0, "Expected off-season factor < 1.0 for SACI in Nov, got {}", f);
    }

    #[test]
    fn test_exdiv_bonus() {
        let svc = SeasonalityService::new();
        let mut reasons = Vec::new();
        // High yield stock in February → ex-div rally
        let f = svc.seasonal_factor("SNTS", "Télécom", 2, Some(7.5), &mut reasons);
        assert!(f > 1.0, "Expected ex-div bonus for SNTS with 7.5% yield in Feb, got {}", f);
        assert!(reasons.iter().any(|r| r.contains("ex-dividende")));
    }

    #[test]
    fn test_score_range() {
        let svc = SeasonalityService::new();
        for month in 1..=12 {
            let (score, _) = svc.score("SACI", "Agriculture", month, Some(5.0));
            assert!((0.0..=100.0).contains(&score), "Score out of range: {}", score);
        }
    }
}
