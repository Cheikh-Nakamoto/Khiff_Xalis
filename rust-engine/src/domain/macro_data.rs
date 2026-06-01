// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

#[derive(Debug, Clone, Default)]
pub struct MacroData {
    pub inflation: Option<f64>,
    pub taux_directeur: Option<f64>,
    pub change_xof_eur: Option<f64>,
    pub cocoa_price: Option<f64>,
    pub oil_price: Option<f64>,
    pub political_stability: Option<f64>,
    pub sovereign_rating: Option<i32>,
}
