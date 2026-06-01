// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

pub mod domain;
pub mod services;
pub mod scanner;

// Re-exports des entities du domain
pub use domain::ticker::{Ticker, PortfolioPosition};
pub use domain::market::MarketDataPoint;
pub use domain::fundamental::FundamentalData;
pub use domain::technical::TechnicalIndicators;
pub use domain::macro_data::MacroData;
pub use domain::signal::{SignalType, ScoringResult};

// Re-exports des services
pub use services::technical::TechnicalAnalysisService;
pub use services::fundamental::FundamentalScoringService;
pub use services::macro_scoring::MacroScoringService;
pub use services::diversification::DiversificationService;
pub use services::signal::SignalGenerationService;

// Re-exports du scanner
pub use scanner::{scan_all_tickers, scan_all_tickers_with_macro};
