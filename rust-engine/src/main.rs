// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

use tonic::{Request, Response, Status};
use tracing_subscriber::EnvFilter;

// Include generated proto code
pub mod signals {
    tonic::include_proto!("brvm.signals");
}

use signals::signal_service_server::{SignalService, SignalServiceServer};
use signals::*;

// Import from lib.rs
use brvm_engine::{
    FundamentalData as EngineFundamentalData, SignalGenerationService, Ticker,
    TechnicalAnalysisService, scan_all_tickers,
};
use std::collections::HashMap;

pub struct SignalServiceImpl {
    engine: SignalGenerationService,
}

#[tonic::async_trait]
impl SignalService for SignalServiceImpl {
    async fn generate_signal(
        &self,
        request: Request<SignalRequest>,
    ) -> Result<Response<SignalResponse>, Status> {
        let req = request.into_inner();

        let ticker = Ticker {
            symbol: req.ticker.clone(),
            name: String::new(),
            country: String::new(),
            sector: String::new(),
        };

        let fundamental = match req.fundamental {
            Some(f) => EngineFundamentalData {
                per: f.per,
                roe: f.roe,
                dividend_yield: f.dividend_yield,
                eps: f.eps,
                book_value_per_share: f.book_value_per_share,
                debt_to_equity: f.debt_to_equity,
                revenue_growth: f.revenue_growth,
                net_profit: f.net_profit,
            },
            None => EngineFundamentalData::default(),
        };

        let result = self.engine.generate_signal(
            ticker,
            &req.prices,
            &req.volumes,
            &req.highs,
            &req.lows,
            &fundamental,
            req.current_price,
        );

        let response = SignalResponse {
            ticker: req.ticker,
            composite_score: result.composite_score,
            signal: result.signal.to_string(),
            confidence: result.confidence,
            reasons: result.reasons,
            technical: Some(TechnicalIndicators {
                sma_20: result.technical_indicators.sma_20,
                sma_50: result.technical_indicators.sma_50,
                ema_12: result.technical_indicators.ema_12,
                ema_26: result.technical_indicators.ema_26,
                rsi_14: result.technical_indicators.rsi_14,
                macd: result.technical_indicators.macd,
                macd_signal: result.technical_indicators.macd_signal,
                bollinger_upper: result.technical_indicators.bollinger_upper,
                bollinger_lower: result.technical_indicators.bollinger_lower,
                atr_14: result.technical_indicators.atr_14,
                volume_sma_20: result.technical_indicators.volume_sma_20,
            }),
            fundamental_scores: result.fundamental_scores,
            technical_score: result.technical_score,
            risk_score: result.risk_score,
        };

        Ok(Response::new(response))
    }

    async fn scan_all_tickers(
        &self,
        request: Request<ScanRequest>,
    ) -> Result<Response<ScanResponse>, Status> {
        let req = request.into_inner();

        let tickers: Vec<Ticker> = req
            .tickers
            .iter()
            .map(|t| Ticker {
                symbol: t.symbol.clone(),
                name: t.name.clone(),
                country: t.country.clone(),
                sector: t.sector.clone(),
            })
            .collect();

        let mut data_map = HashMap::new();
        let mut fund_map = HashMap::new();
        let mut current_prices = HashMap::new();

        for t in &req.tickers {
            data_map.insert(t.symbol.clone(), (t.prices.clone(), t.volumes.clone(), t.highs.clone(), t.lows.clone()));
            current_prices.insert(t.symbol.clone(), t.current_price);

            if let Some(ref f) = t.fundamental {
                fund_map.insert(
                    t.symbol.clone(),
                    EngineFundamentalData {
                        per: f.per,
                        roe: f.roe,
                        dividend_yield: f.dividend_yield,
                        eps: f.eps,
                        book_value_per_share: f.book_value_per_share,
                        debt_to_equity: f.debt_to_equity,
                        revenue_growth: f.revenue_growth,
                        net_profit: f.net_profit,
                    },
                );
            }
        }

        let results = scan_all_tickers(tickers, data_map, fund_map, current_prices);

        let signals: Vec<SignalResponse> = results
            .into_iter()
            .map(|r| SignalResponse {
                ticker: r.ticker.symbol,
                composite_score: r.composite_score,
                signal: r.signal.to_string(),
                confidence: r.confidence,
                reasons: r.reasons,
                technical: Some(TechnicalIndicators {
                    sma_20: r.technical_indicators.sma_20,
                    sma_50: r.technical_indicators.sma_50,
                    ema_12: r.technical_indicators.ema_12,
                    ema_26: r.technical_indicators.ema_26,
                    rsi_14: r.technical_indicators.rsi_14,
                    macd: r.technical_indicators.macd,
                    macd_signal: r.technical_indicators.macd_signal,
                    bollinger_upper: r.technical_indicators.bollinger_upper,
                    bollinger_lower: r.technical_indicators.bollinger_lower,
                    atr_14: r.technical_indicators.atr_14,
                    volume_sma_20: r.technical_indicators.volume_sma_20,
                }),
                fundamental_scores: r.fundamental_scores,
                technical_score: r.technical_score,
                risk_score: r.risk_score,
            })
            .collect();

        let total = signals.len() as i32;

        Ok(Response::new(ScanResponse {
            signals,
            total_scanned: total,
            total_signals: total,
        }))
    }

    async fn health_check(
        &self,
        _request: Request<HealthRequest>,
    ) -> Result<Response<HealthResponse>, Status> {
        Ok(Response::new(HealthResponse {
            status: "ok".to_string(),
            version: env!("CARGO_PKG_VERSION").to_string(),
        }))
    }
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();

    let addr = "0.0.0.0:50051".parse()?;
    let service = SignalServiceImpl {
        engine: SignalGenerationService::new(),
    };

    tracing::info!("BRVM Rust Engine gRPC server starting on {}", addr);

    tonic::transport::Server::builder()
        .add_service(SignalServiceServer::new(service))
        .serve(addr)
        .await?;

    Ok(())
}
