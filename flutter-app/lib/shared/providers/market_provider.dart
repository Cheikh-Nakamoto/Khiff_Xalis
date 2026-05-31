import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/network/api_client.dart';
import '../models/market_data.dart';

// ─────────────────────────────────────────
// TICKERS
// ─────────────────────────────────────────

/// Fetches all available tickers from the API.
final tickersProvider = FutureProvider<List<TickerInfo>>((ref) async {
  final api = ref.read(apiClientProvider);
  try {
    final response = await api.getTickers();
    final data = response.data as Map<String, dynamic>;
    final list = data['data'] as List<dynamic>;
    return list
        .map((e) => TickerInfo.fromJson(e as Map<String, dynamic>))
        .toList();
  } catch (_) {
    // Mock fallback
    return _mockTickers;
  }
});

// ─────────────────────────────────────────
// MARKET DATA (OHLCV) per ticker
// ─────────────────────────────────────────

/// Fetches OHLCV market data for a given ticker.
final marketDataProvider =
    FutureProvider.family<List<MarketData>, String>((ref, ticker) async {
  final api = ref.read(apiClientProvider);
  try {
    final response = await api.getMarketData(ticker);
    final data = response.data as Map<String, dynamic>;
    final list = data['data'] as List<dynamic>;
    return list
        .map((e) => MarketData.fromJson(e as Map<String, dynamic>))
        .toList();
  } catch (_) {
    // Mock fallback
    return _mockMarketData;
  }
});

/// Fetches the latest price for a ticker.
final latestPriceProvider =
    FutureProvider.family<MarketData?, String>((ref, ticker) async {
  final api = ref.read(apiClientProvider);
  try {
    final response = await api.getLatestData(ticker);
    return MarketData.fromJson(response.data as Map<String, dynamic>);
  } catch (_) {
    return null;
  }
});

// ─────────────────────────────────────────
// FUNDAMENTALS
// ─────────────────────────────────────────

/// Fetches fundamental data for a ticker.
final fundamentalsProvider =
    FutureProvider.family<FundamentalData?, String>((ref, ticker) async {
  final api = ref.read(apiClientProvider);
  try {
    final response = await api.getFundamentals(ticker);
    return FundamentalData.fromJson(response.data as Map<String, dynamic>);
  } catch (_) {
    return null;
  }
});

// ─────────────────────────────────────────
// SECTOR HEATMAP DATA
// ─────────────────────────────────────────

/// Aggregated sector performance for the heatmap.
/// Computed from tickers + latest prices.
final sectorHeatmapProvider = FutureProvider<List<SectorHeatEntry>>((ref) async {
  final tickers = await ref.watch(tickersProvider.future);

  // Group tickers by sector
  final sectorMap = <String, List<TickerInfo>>{};
  for (final t in tickers) {
    sectorMap.putIfAbsent(t.sector, () => []).add(t);
  }

  // For now, return mock sector changes (real impl would fetch latest prices)
  return sectorMap.entries.map((entry) {
    final change = _mockSectorChanges[entry.key] ?? 0.0;
    return SectorHeatEntry(
      sector: entry.key,
      tickerCount: entry.value.length,
      changePercent: change,
    );
  }).toList();
});

class SectorHeatEntry {
  const SectorHeatEntry({
    required this.sector,
    required this.tickerCount,
    required this.changePercent,
  });

  final String sector;
  final int tickerCount;
  final double changePercent;
}

// ─────────────────────────────────────────
// MOCK DATA (fallback when API is down)
// ─────────────────────────────────────────

const _mockTickers = [
  TickerInfo(symbol: 'SNTS', name: 'Sonatel', country: 'Sénégal', sector: 'Télécom'),
  TickerInfo(symbol: 'ORAC', name: 'Orange CI', country: "Côte d'Ivoire", sector: 'Télécom'),
  TickerInfo(symbol: 'SGBC', name: 'SGB CI', country: "Côte d'Ivoire", sector: 'Finance'),
  TickerInfo(symbol: 'ECOC', name: 'Ecobank CI', country: "Côte d'Ivoire", sector: 'Finance'),
  TickerInfo(symbol: 'PALC', name: 'PalmCI', country: "Côte d'Ivoire", sector: 'Agriculture'),
  TickerInfo(symbol: 'SPHC', name: 'SAPH CI', country: "Côte d'Ivoire", sector: 'Agriculture'),
  TickerInfo(symbol: 'SMB', name: 'SMB CI', country: "Côte d'Ivoire", sector: 'Industrie'),
  TickerInfo(symbol: 'SOGC', name: 'SOGB CI', country: "Côte d'Ivoire", sector: 'Industrie'),
  TickerInfo(symbol: 'SLBC', name: 'Solibra CI', country: "Côte d'Ivoire", sector: 'Consommation'),
  TickerInfo(symbol: 'TTLC', name: 'Total CI', country: "Côte d'Ivoire", sector: 'Distribution'),
  TickerInfo(symbol: 'TTLS', name: 'Total SN', country: 'Sénégal', sector: 'Distribution'),
  TickerInfo(symbol: 'UNLC', name: 'Unilever CI', country: "Côte d'Ivoire", sector: 'Consommation'),
  TickerInfo(symbol: 'NEST', name: 'Nestlé CI', country: "Côte d'Ivoire", sector: 'Consommation'),
  TickerInfo(symbol: 'NSIA', name: 'NSIA Banque', country: "Côte d'Ivoire", sector: 'Finance'),
  TickerInfo(symbol: 'ETI', name: 'ETI TG', country: 'Togo', sector: 'Finance'),
  TickerInfo(symbol: 'ORGT', name: 'Oragroup TG', country: 'Togo', sector: 'Finance'),
  TickerInfo(symbol: 'ONTBF', name: 'Onatel BF', country: 'Burkina Faso', sector: 'Télécom'),
  TickerInfo(symbol: 'LNB', name: 'Loterie Nationale Bénin', country: 'Bénin', sector: 'Services publics'),
];

final _mockMarketData = List.generate(
  90,
  (i) => MarketData(
    date: DateTime(2026, 1, 1).add(Duration(days: i)).toString().substring(0, 10),
    open: 25000 + (i * 50) + (i % 3 == 0 ? 200 : -100),
    high: 25500 + (i * 50),
    low: 24500 + (i * 50),
    close: 25200 + (i * 55) + (i % 2 == 0 ? 150 : -80),
    volume: 1500 + (i * 30),
    dailyReturn: i % 2 == 0 ? 0.8 : -0.4,
  ),
);

const _mockSectorChanges = {
  'Télécom': 2.4,
  'Finance': 1.8,
  'Industrie': 0.5,
  'Agriculture': 3.1,
  'Consommation': -0.2,
  'Distribution': 0.8,
  'Services publics': 0.1,
};
