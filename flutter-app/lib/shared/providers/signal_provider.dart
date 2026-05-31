import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/network/api_client.dart';
import '../models/signal.dart';

/// Fetches all signals from the scan endpoint.
///
/// Optionally filter by [signalType]: BUY, STRONG_BUY, SELL, STRONG_SELL, HOLD.
final signalsProvider =
    FutureProvider.family<List<Signal>, String?>((ref, signalType) async {
  final api = ref.read(apiClientProvider);
  try {
    final response = await api.scanSignals(
      minScore: 50,
      signalType: signalType,
    );
    final data = response.data as Map<String, dynamic>;
    final list = data['data'] as List<dynamic>;
    return list
        .map((e) => Signal.fromJson(e as Map<String, dynamic>))
        .toList();
  } catch (_) {
    // Mock fallback
    if (signalType != null) {
      return _mockSignals
          .where((s) => s.signal == signalType)
          .toList();
    }
    return _mockSignals;
  }
});

/// Fetches a signal for a specific ticker.
final tickerSignalProvider =
    FutureProvider.family<Signal?, String>((ref, ticker) async {
  final api = ref.read(apiClientProvider);
  try {
    final response = await api.getSignal(ticker);
    final data = response.data as Map<String, dynamic>;
    if (data.containsKey('signal')) {
      return Signal.fromJson(data);
    }
    // Cached response format
    final signalData = data['signal'];
    if (signalData is Map<String, dynamic>) {
      return Signal.fromJson(signalData);
    }
    return null;
  } catch (_) {
    // Mock fallback
    try {
      return _mockSignals.firstWhere((s) => s.tickerSymbol == ticker);
    } catch (_) {
      return null;
    }
  }
});

/// Top signals for the dashboard (highest scores).
final topSignalsProvider = FutureProvider<List<Signal>>((ref) async {
  final signals = await ref.watch(signalsProvider(null).future);
  final sorted = List<Signal>.from(signals)
    ..sort((a, b) => b.compositeScore.compareTo(a.compositeScore));
  return sorted.take(5).toList();
});

// ─────────────────────────────────────────
// MOCK DATA
// ─────────────────────────────────────────

const _mockSignals = [
  Signal(
    tickerSymbol: 'SIBC',
    tickerName: 'SIB CI',
    compositeScore: 88.8,
    signal: 'STRONG_BUY',
    confidence: 0.92,
    reasons: ['RSI en zone de survente', 'PER sous-évalué', 'Dividende élevé'],
  ),
  Signal(
    tickerSymbol: 'SNTS',
    tickerName: 'Sonatel SN',
    compositeScore: 87.4,
    signal: 'BUY',
    confidence: 0.88,
    reasons: ['Tendance haussière MACD', 'Volume en augmentation'],
  ),
  Signal(
    tickerSymbol: 'SPHC',
    tickerName: 'SAPH CI',
    compositeScore: 84.1,
    signal: 'STRONG_BUY',
    confidence: 0.85,
    reasons: ['Cassure résistance', 'Fondamentaux solides'],
  ),
  Signal(
    tickerSymbol: 'PALC',
    tickerName: 'PalmCI',
    compositeScore: 65.7,
    signal: 'BUY',
    confidence: 0.70,
    reasons: ['Secteur agricole en hausse'],
  ),
  Signal(
    tickerSymbol: 'ORGT',
    tickerName: 'Oragroup TG',
    compositeScore: 65.6,
    signal: 'BUY',
    confidence: 0.68,
    reasons: ['PER attractif', 'ROE élevé'],
  ),
  Signal(
    tickerSymbol: 'ETI',
    tickerName: 'ETI TG',
    compositeScore: 55.0,
    signal: 'HOLD',
    confidence: 0.55,
    reasons: ['Consolidation en cours'],
  ),
  Signal(
    tickerSymbol: 'UNXC',
    tickerName: 'Uniwax CI',
    compositeScore: 38.9,
    signal: 'SELL',
    confidence: 0.72,
    reasons: ['RSI en surachat', 'Dette élevée'],
  ),
];
