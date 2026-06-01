// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

use std::collections::HashMap;
use rayon::prelude::*;
use crate::domain::*;
use crate::services::SignalGenerationService;

type MarketDataBundle = (Vec<f64>, Vec<i64>, Vec<f64>, Vec<f64>);

pub fn scan_all_tickers(
    tickers: Vec<Ticker>,
    data_map: HashMap<String, MarketDataBundle>,
    fund_map: HashMap<String, FundamentalData>,
    current_prices: HashMap<String, f64>,
) -> Vec<ScoringResult> {
    scan_all_tickers_with_macro(tickers, data_map, fund_map, current_prices, HashMap::new(), &[])
}

pub fn scan_all_tickers_with_macro(
    tickers: Vec<Ticker>,
    data_map: HashMap<String, MarketDataBundle>,
    fund_map: HashMap<String, FundamentalData>,
    current_prices: HashMap<String, f64>,
    macro_map: HashMap<String, MacroData>,
    user_portfolio: &[PortfolioPosition],
) -> Vec<ScoringResult> {
    let engine = SignalGenerationService::new();

    tickers
        .into_par_iter()
        .filter_map(|ticker| {
            let symbol = ticker.symbol.clone();
            let (closes, volumes, highs, lows) = data_map.get(&symbol)?;
            let fundamental = fund_map.get(&symbol).cloned().unwrap_or_default();
            let current_price = *current_prices.get(&symbol)?;
            let macro_data = macro_map.get(&symbol).cloned().unwrap_or_default();

            Some(engine.generate_signal_with_macro(
                ticker,
                closes,
                volumes,
                highs,
                lows,
                &fundamental,
                current_price,
                &macro_data,
                user_portfolio,
            ))
        })
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

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
            println!("{}: {} (Score: {:.1})", r.ticker.symbol, r.signal, r.composite_score);
        }
    }
}
