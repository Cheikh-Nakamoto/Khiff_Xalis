/// Market data point representing OHLCV candlestick data.
class MarketData {
  const MarketData({
    required this.date,
    required this.open,
    required this.high,
    required this.low,
    required this.close,
    required this.volume,
    this.dailyReturn = 0.0,
    this.source = '',
  });

  final String date;
  final double open;
  final double high;
  final double low;
  final double close;
  final int volume;
  final double dailyReturn;
  final String source;

  factory MarketData.fromJson(Map<String, dynamic> json) {
    return MarketData(
      date: json['date'] as String? ?? '',
      open: (json['open'] as num?)?.toDouble() ?? 0.0,
      high: (json['high'] as num?)?.toDouble() ?? 0.0,
      low: (json['low'] as num?)?.toDouble() ?? 0.0,
      close: (json['close'] as num?)?.toDouble() ?? 0.0,
      volume: (json['volume'] as num?)?.toInt() ?? 0,
      dailyReturn: (json['daily_return'] as num?)?.toDouble() ?? 0.0,
      source: json['source'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() => {
        'date': date,
        'open': open,
        'high': high,
        'low': low,
        'close': close,
        'volume': volume,
        'daily_return': dailyReturn,
        'source': source,
      };
}

/// Ticker metadata (symbol, name, country, sector).
class TickerInfo {
  const TickerInfo({
    required this.symbol,
    required this.name,
    required this.country,
    required this.sector,
  });

  final String symbol;
  final String name;
  final String country;
  final String sector;

  factory TickerInfo.fromJson(Map<String, dynamic> json) {
    return TickerInfo(
      symbol: json['symbol'] as String? ?? '',
      name: json['name'] as String? ?? '',
      country: json['country'] as String? ?? '',
      sector: json['sector'] as String? ?? '',
    );
  }
}

/// Fundamental data for a ticker.
class FundamentalData {
  const FundamentalData({
    required this.ticker,
    this.per,
    this.roe,
    this.dividendYield,
    this.eps,
    this.bookValuePerShare,
    this.debtToEquity,
    this.revenueGrowth,
    this.netProfit,
    this.updatedAt,
  });

  final String ticker;
  final double? per;
  final double? roe;
  final double? dividendYield;
  final double? eps;
  final double? bookValuePerShare;
  final double? debtToEquity;
  final double? revenueGrowth;
  final double? netProfit;
  final String? updatedAt;

  factory FundamentalData.fromJson(Map<String, dynamic> json) {
    return FundamentalData(
      ticker: json['ticker'] as String? ?? '',
      per: (json['per'] as num?)?.toDouble(),
      roe: (json['roe'] as num?)?.toDouble(),
      dividendYield: (json['dividend_yield'] as num?)?.toDouble(),
      eps: (json['eps'] as num?)?.toDouble(),
      bookValuePerShare: (json['book_value_per_share'] as num?)?.toDouble(),
      debtToEquity: (json['debt_to_equity'] as num?)?.toDouble(),
      revenueGrowth: (json['revenue_growth'] as num?)?.toDouble(),
      netProfit: (json['net_profit'] as num?)?.toDouble(),
      updatedAt: json['updated_at'] as String?,
    );
  }
}
