/// Portfolio position for a single ticker.
class PortfolioPosition {
  const PortfolioPosition({
    required this.ticker,
    required this.quantity,
    required this.avgBuyPrice,
    required this.currentPrice,
    required this.marketValue,
    required this.unrealizedPnL,
    required this.unrealizedPnLPct,
    this.name = '',
  });

  final String ticker;
  final String name;
  final int quantity;
  final double avgBuyPrice;
  final double currentPrice;
  final double marketValue;
  final double unrealizedPnL;
  final double unrealizedPnLPct;

  bool get isProfit => unrealizedPnL >= 0;

  factory PortfolioPosition.fromJson(Map<String, dynamic> json) {
    return PortfolioPosition(
      ticker: json['ticker'] as String? ?? '',
      quantity: (json['quantity'] as num?)?.toInt() ?? 0,
      avgBuyPrice: (json['avg_buy_price'] as num?)?.toDouble() ?? 0.0,
      currentPrice: (json['current_price'] as num?)?.toDouble() ?? 0.0,
      marketValue: (json['market_value'] as num?)?.toDouble() ?? 0.0,
      unrealizedPnL: (json['unrealized_pnl'] as num?)?.toDouble() ?? 0.0,
      unrealizedPnLPct:
          (json['unrealized_pnl_pct'] as num?)?.toDouble() ?? 0.0,
      name: json['name'] as String? ?? '',
    );
  }
}

/// Portfolio summary with all positions.
class PortfolioSummary {
  const PortfolioSummary({
    required this.totalValue,
    required this.totalPnL,
    required this.positions,
  });

  final double totalValue;
  final double totalPnL;
  final List<PortfolioPosition> positions;

  double get totalPnLPct =>
      totalValue > 0 ? totalPnL / (totalValue - totalPnL) * 100 : 0;

  factory PortfolioSummary.fromJson(Map<String, dynamic> json) {
    return PortfolioSummary(
      totalValue: (json['total_value'] as num?)?.toDouble() ?? 0.0,
      totalPnL: (json['total_pnl'] as num?)?.toDouble() ?? 0.0,
      positions: (json['positions'] as List<dynamic>?)
              ?.map((e) =>
                  PortfolioPosition.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );
  }

  /// Empty portfolio for mock fallback.
  factory PortfolioSummary.empty() => const PortfolioSummary(
        totalValue: 0,
        totalPnL: 0,
        positions: [],
      );
}
