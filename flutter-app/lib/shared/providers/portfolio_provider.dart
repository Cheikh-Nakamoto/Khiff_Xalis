import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/network/api_client.dart';
import '../models/portfolio_position.dart';

/// Portfolio state notifier.
class PortfolioNotifier extends StateNotifier<AsyncValue<PortfolioSummary>> {
  PortfolioNotifier(this._api) : super(const AsyncValue.loading()) {
    loadPortfolio();
  }

  final ApiClient _api;

  /// Fetch portfolio from the API.
  Future<void> loadPortfolio() async {
    state = const AsyncValue.loading();
    try {
      final response = await _api.getPortfolio();
      final summary = PortfolioSummary.fromJson(
        response.data as Map<String, dynamic>,
      );
      state = AsyncValue.data(summary);
    } catch (e, st) {
      // Mock fallback
      state = AsyncValue.data(_mockPortfolio);
      // Optionally log: print('Portfolio load failed, using mock: $e');
    }
  }

  /// Refresh portfolio data.
  Future<void> refresh() => loadPortfolio();
}

/// Provider for the portfolio state.
final portfolioProvider =
    StateNotifierProvider<PortfolioNotifier, AsyncValue<PortfolioSummary>>(
        (ref) {
  final api = ref.read(apiClientProvider);
  return PortfolioNotifier(api);
});

/// Convenience provider for total portfolio value.
final portfolioValueProvider = Provider<double>((ref) {
  final portfolio = ref.watch(portfolioProvider);
  return portfolio.valueOrNull?.totalValue ?? 0.0;
});

/// Convenience provider for total P&L.
final portfolioPnLProvider = Provider<double>((ref) {
  final portfolio = ref.watch(portfolioProvider);
  return portfolio.valueOrNull?.totalPnL ?? 0.0;
});

// ─────────────────────────────────────────
// MOCK DATA
// ─────────────────────────────────────────

const _mockPortfolio = PortfolioSummary(
  totalValue: 3975000,
  totalPnL: 125000,
  positions: [
    PortfolioPosition(
      ticker: 'SNTS',
      name: 'Sonatel',
      quantity: 100,
      avgBuyPrice: 25000,
      currentPrice: 26000,
      marketValue: 2600000,
      unrealizedPnL: 100000,
      unrealizedPnLPct: 4.0,
    ),
    PortfolioPosition(
      ticker: 'SGBC',
      name: 'SGB CI',
      quantity: 50,
      avgBuyPrice: 27000,
      currentPrice: 27500,
      marketValue: 1375000,
      unrealizedPnL: 25000,
      unrealizedPnLPct: 1.85,
    ),
    PortfolioPosition(
      ticker: 'ORAC',
      name: 'Orange CI',
      quantity: 200,
      avgBuyPrice: 14500,
      currentPrice: 14650,
      marketValue: 2930000,
      unrealizedPnL: 30000,
      unrealizedPnLPct: 1.03,
    ),
    PortfolioPosition(
      ticker: 'PALC',
      name: 'PalmCI',
      quantity: 300,
      avgBuyPrice: 9600,
      currentPrice: 9515,
      marketValue: 2854500,
      unrealizedPnL: -25500,
      unrealizedPnLPct: -0.89,
    ),
    PortfolioPosition(
      ticker: 'SPHC',
      name: 'SAPH CI',
      quantity: 150,
      avgBuyPrice: 8000,
      currentPrice: 7880,
      marketValue: 1182000,
      unrealizedPnL: -18000,
      unrealizedPnLPct: -1.50,
    ),
  ],
);
