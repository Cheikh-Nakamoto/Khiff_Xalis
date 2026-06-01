// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

use std::collections::HashMap;
use crate::domain::FundamentalData;

pub struct FundamentalScoringService;

impl Default for FundamentalScoringService {
    fn default() -> Self {
        Self::new()
    }
}

impl FundamentalScoringService {
    pub fn new() -> Self {
        Self
    }

    pub fn score(&self, fundamental: &FundamentalData) -> (HashMap<String, f64>, Vec<String>) {
        let mut scores = HashMap::new();
        let mut reasons = Vec::new();

        let per_score = match fundamental.per {
            Some(p) if p < 8.0 => { reasons.push(format!("PER très attractif: {:.1}", p)); 95.0 }
            Some(p) if p < 12.0 => { reasons.push(format!("PER attractif: {:.1}", p)); 80.0 }
            Some(p) if p < 18.0 => 60.0,
            Some(p) if p < 25.0 => { reasons.push(format!("PER élevé: {:.1}", p)); 40.0 }
            Some(p) => { reasons.push(format!("PER très élevé: {:.1}", p)); 20.0 }
            None => 50.0,
        };
        scores.insert("per".to_string(), per_score);

        let roe_score = match fundamental.roe {
            Some(r) if r > 25.0 => { reasons.push(format!("ROE excellent: {:.1}%", r)); 95.0 }
            Some(r) if r > 20.0 => { reasons.push(format!("ROE très bon: {:.1}%", r)); 85.0 }
            Some(r) if r > 15.0 => 70.0,
            Some(r) if r > 10.0 => 50.0,
            Some(r) => { reasons.push(format!("ROE faible: {:.1}%", r)); 30.0 }
            None => 50.0,
        };
        scores.insert("roe".to_string(), roe_score);

        let div_score = match fundamental.dividend_yield {
            Some(d) if d > 7.0 => { reasons.push(format!("Yield exceptionnel: {:.1}%", d)); 95.0 }
            Some(d) if d > 5.0 => { reasons.push(format!("Yield attractif: {:.1}%", d)); 80.0 }
            Some(d) if d > 3.0 => 60.0,
            Some(d) if d > 1.0 => 40.0,
            Some(_) => 20.0,
            None => { reasons.push("Pas de dividende".to_string()); 30.0 }
        };
        scores.insert("dividend".to_string(), div_score);

        let debt_score = match fundamental.debt_to_equity {
            Some(d) if d < 0.5 => 90.0,
            Some(d) if d < 1.0 => 70.0,
            Some(d) if d < 1.5 => 50.0,
            Some(d) => { reasons.push(format!("Dette élevée: {:.2}", d)); 30.0 }
            None => 50.0,
        };
        scores.insert("debt".to_string(), debt_score);

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

#[cfg(test)]
mod tests {
    use super::*;

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
}
