use criterion::{black_box, criterion_group, criterion_main, Criterion};
use brvm_engine::{TechnicalAnalysisService, SignalGenerationService, Ticker, FundamentalData};

fn bench_sma(c: &mut Criterion) {
    let service = TechnicalAnalysisService::new();
    let prices: Vec<f64> = (0..1000).map(|i| 100.0 + i as f64 * 0.5).collect();

    let highs: Vec<f64> = prices.iter().map(|p| p + 1.0).collect();
    let lows: Vec<f64> = prices.iter().map(|p| p - 1.0).collect();

    c.bench_function("sma_20", |b| {
        b.iter(|| service.calculate(black_box(&prices), black_box(&vec![5000; 1000]), black_box(&highs), black_box(&lows)))
    });
}

fn bench_signal_generation(c: &mut Criterion) {
    let service = SignalGenerationService::new();
    let ticker = Ticker {
        symbol: "SNTS".to_string(),
        name: "Sonatel".to_string(),
        country: "Sénégal".to_string(),
        sector: "Télécom".to_string(),
    };
    let prices: Vec<f64> = (0..60).map(|i| 25000.0 + i as f64 * 50.0).collect();
    let volumes: Vec<i64> = vec![5000; 60];
    let highs: Vec<f64> = prices.iter().map(|p| p + 100.0).collect();
    let lows: Vec<f64> = prices.iter().map(|p| p - 100.0).collect();
    let fund = FundamentalData {
        per: Some(12.5),
        roe: Some(28.5),
        dividend_yield: Some(6.5),
        ..Default::default()
    };

    c.bench_function("signal_generation", |b| {
        b.iter(|| service.generate_signal(
            black_box(ticker.clone()),
            black_box(&prices),
            black_box(&volumes),
            black_box(&highs),
            black_box(&lows),
            black_box(&fund),
            black_box(26000.0),
        ))
    });
}

criterion_group!(benches, bench_sma, bench_signal_generation);
criterion_main!(benches);
