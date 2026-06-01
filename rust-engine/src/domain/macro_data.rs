// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

#[derive(Debug, Clone, Default)]
pub struct MacroData {
    pub inflation: Option<f64>,
    pub taux_directeur: Option<f64>,
    /// XOF/EUR exchange rate. Fixed peg at 655.957 but monitored for deviation.
    pub change_xof_eur: Option<f64>,
    /// XOF/USD exchange rate. Floats freely — key for export-sector scoring.
    pub change_xof_usd: Option<f64>,
    pub cocoa_price: Option<f64>,
    pub oil_price: Option<f64>,
    /// Additional UEMOA commodity prices (cajou, or, caoutchouc, huile de palme).
    pub cashew_price: Option<f64>,
    pub gold_price: Option<f64>,
    pub rubber_price: Option<f64>,
    pub palm_oil_price: Option<f64>,
    /// WGI political stability score: -2.5 (unstable) to +2.5 (stable).
    pub political_stability: Option<f64>,
    /// Political crisis freeze flag: true = suspend all BUY signals for this country.
    pub political_crisis: bool,
    pub sovereign_rating: Option<i32>,
    /// Per-stock commodity beta: regression coefficient stock vs primary commodity.
    /// Pre-computed by go-collector over a 60-day rolling window.
    pub commodity_beta: Option<f64>,
}
