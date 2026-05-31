/// Trading signal result from the scoring engine.
class Signal {
  const Signal({
    required this.tickerSymbol,
    required this.tickerName,
    required this.compositeScore,
    required this.signal,
    this.confidence = 0.0,
    this.reasons = const [],
    this.technicalIndicators = const {},
    this.fundamentalScores = const {},
    this.timestamp,
  });

  final String tickerSymbol;
  final String tickerName;
  final double compositeScore;
  final String signal; // BUY, STRONG_BUY, SELL, STRONG_SELL, HOLD
  final double confidence;
  final List<String> reasons;
  final Map<String, dynamic> technicalIndicators;
  final Map<String, double> fundamentalScores;
  final String? timestamp;

  /// Whether this is a buy-type signal.
  bool get isBuy => signal == 'BUY' || signal == 'STRONG_BUY';

  /// Whether this is a sell-type signal.
  bool get isSell => signal == 'SELL' || signal == 'STRONG_SELL';

  /// Human-readable label in French.
  String get labelFr {
    switch (signal) {
      case 'STRONG_BUY':
        return 'ACHAT FORT';
      case 'BUY':
        return 'ACHAT';
      case 'STRONG_SELL':
        return 'VENTE FORTE';
      case 'SELL':
        return 'VENTE';
      default:
        return 'NEUTRE';
    }
  }

  factory Signal.fromJson(Map<String, dynamic> json) {
    final ticker = json['ticker'];
    String symbol = '';
    String name = '';
    if (ticker is Map<String, dynamic>) {
      symbol = ticker['symbol'] as String? ?? '';
      name = ticker['name'] as String? ?? symbol;
    }

    return Signal(
      tickerSymbol: symbol,
      tickerName: name,
      compositeScore: (json['composite_score'] as num?)?.toDouble() ?? 0.0,
      signal: json['signal'] as String? ?? 'HOLD',
      confidence: (json['confidence'] as num?)?.toDouble() ?? 0.0,
      reasons: (json['reasons'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          const [],
      technicalIndicators:
          (json['technical_indicators'] as Map<String, dynamic>?) ?? const {},
      fundamentalScores: (json['fundamental_scores'] as Map<String, dynamic>?)
              ?.map((k, v) => MapEntry(k, (v as num).toDouble())) ??
          const {},
      timestamp: json['timestamp'] as String?,
    );
  }
}
