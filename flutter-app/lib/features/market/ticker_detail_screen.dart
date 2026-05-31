import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fl_chart/fl_chart.dart';
import '../../shared/providers/market_provider.dart';
import '../../shared/providers/signal_provider.dart';
import '../../shared/models/market_data.dart';
import '../../shared/models/signal.dart';

/// Detail screen for a single ticker.
///
/// Shows price, OHLCV chart, technical indicators, fundamental data,
/// and the current trading signal.
class TickerDetailScreen extends ConsumerWidget {
  final String ticker;
  const TickerDetailScreen({super.key, required this.ticker});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final marketDataAsync = ref.watch(marketDataProvider(ticker));
    final signalAsync = ref.watch(tickerSignalProvider(ticker));
    final fundamentalsAsync = ref.watch(fundamentalsProvider(ticker));

    return Scaffold(
      appBar: AppBar(
        title: Text(ticker, style: const TextStyle(fontWeight: FontWeight.bold)),
        actions: [
          IconButton(icon: const Icon(Icons.star_border), onPressed: () {}),
          IconButton(icon: const Icon(Icons.notifications_none), onPressed: () {}),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Price Header
            _buildPriceHeader(signalAsync),
            const SizedBox(height: 24),

            // OHLCV Chart
            Text(
              'Graphique OHLCV',
              style: Theme.of(context)
                  .textTheme
                  .titleLarge
                  ?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            _buildChart(marketDataAsync),
            const SizedBox(height: 24),

            // Technical Indicators
            Text(
              'Indicateurs Techniques',
              style: Theme.of(context)
                  .textTheme
                  .titleLarge
                  ?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            _buildIndicatorGrid(signalAsync),
            const SizedBox(height: 24),

            // Fundamental Data
            Text(
              'Donnees Fondamentales',
              style: Theme.of(context)
                  .textTheme
                  .titleLarge
                  ?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            _buildFundamentalTable(fundamentalsAsync),
            const SizedBox(height: 24),

            // Action Buttons
            _buildActionButtons(),
          ],
        ),
      ),
    );
  }

  Widget _buildPriceHeader(AsyncValue<Signal?> signalAsync) {
    final signal = signalAsync.valueOrNull;
    final signalLabel = signal?.labelFr ?? 'N/A';
    final isBuy = signal?.isBuy ?? true;
    final color = isBuy ? Colors.green : (signal?.isSell ?? false) ? Colors.red : Colors.orange;

    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              '26 000 FCFA',
              style: TextStyle(fontSize: 36, fontWeight: FontWeight.bold),
            ),
            Row(
              children: [
                const Icon(Icons.arrow_upward, color: Colors.green, size: 16),
                const Text(
                  ' +1.24%',
                  style: TextStyle(
                    color: Colors.green,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Text(
                  ' (+320 FCFA)',
                  style: TextStyle(color: Colors.grey.shade600),
                ),
              ],
            ),
          ],
        ),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          decoration: BoxDecoration(
            color: color,
            borderRadius: BorderRadius.circular(20),
          ),
          child: Text(
            signalLabel,
            style: const TextStyle(
              color: Colors.white,
              fontWeight: FontWeight.bold,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildChart(AsyncValue<List<MarketData>> marketDataAsync) {
    return marketDataAsync.when(
      loading: () => Container(
        height: 250,
        decoration: BoxDecoration(
          color: Colors.grey.shade100,
          borderRadius: BorderRadius.circular(16),
        ),
        child: const Center(child: CircularProgressIndicator()),
      ),
      error: (_, __) => Container(
        height: 250,
        decoration: BoxDecoration(
          color: Colors.grey.shade100,
          borderRadius: BorderRadius.circular(16),
        ),
        child: const Center(child: Text('Donnees indisponibles')),
      ),
      data: (data) {
        if (data.isEmpty) {
          return Container(
            height: 250,
            decoration: BoxDecoration(
              color: Colors.grey.shade100,
              borderRadius: BorderRadius.circular(16),
            ),
            child: const Center(child: Text('Aucune donnee')),
          );
        }

        // Take last 30 points for the chart
        final chartData = data.length > 30 ? data.sublist(data.length - 30) : data;
        final spots = <FlSpot>[];
        for (int i = 0; i < chartData.length; i++) {
          spots.add(FlSpot(i.toDouble(), chartData[i].close));
        }

        final minY = spots.map((s) => s.y).reduce((a, b) => a < b ? a : b);
        final maxY = spots.map((s) => s.y).reduce((a, b) => a > b ? a : b);

        return Container(
          height: 250,
          decoration: BoxDecoration(
            color: Colors.grey.shade50,
            borderRadius: BorderRadius.circular(16),
          ),
          padding: const EdgeInsets.all(16),
          child: LineChart(
            LineChartData(
              gridData: const FlGridData(show: false),
              titlesData: const FlTitlesData(show: false),
              borderData: FlBorderData(show: false),
              minY: minY * 0.99,
              maxY: maxY * 1.01,
              lineBarsData: [
                LineChartBarData(
                  spots: spots,
                  isCurved: true,
                  color: Colors.blue.shade700,
                  barWidth: 2,
                  dotData: const FlDotData(show: false),
                  belowBarData: BarAreaData(
                    show: true,
                    color: Colors.blue.shade100.withOpacity(0.3),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildIndicatorGrid(AsyncValue<Signal?> signalAsync) {
    final signal = signalAsync.valueOrNull;

    // Extract indicators from signal or use defaults
    final rsi = signal?.technicalIndicators['rsi']?.toString() ?? '38.0';
    final macd = signal?.technicalIndicators['macd']?.toString() ?? '+120';

    final indicators = [
      ('RSI(14)', rsi, 'Survente', Colors.green),
      ('SMA 20', '25 500', 'Au-dessus', Colors.green),
      ('SMA 50', '25 000', 'Au-dessus', Colors.green),
      ('MACD', macd, 'Haussier', Colors.green),
      ('BB Upper', '27 000', 'Resistance', Colors.orange),
      ('BB Lower', '24 000', 'Support', Colors.blue),
      ('ATR(14)', '520', 'Modere', Colors.grey),
      ('Volume', '2 900', 'Normal', Colors.grey),
    ];

    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        childAspectRatio: 1.8,
        crossAxisSpacing: 8,
        mainAxisSpacing: 8,
      ),
      itemCount: indicators.length,
      itemBuilder: (context, index) {
        final (label, value, status, color) = indicators[index];
        return Card(
          elevation: 1,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  label,
                  style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
                ),
                Text(
                  value,
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: color.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    status,
                    style: TextStyle(
                      fontSize: 11,
                      color: color,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildFundamentalTable(AsyncValue<dynamic> fundamentalsAsync) {
    return fundamentalsAsync.when(
      loading: () => const Card(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: Center(child: CircularProgressIndicator()),
        ),
      ),
      error: (_, __) => _buildFundamentalTableMock(),
      data: (fundamentals) {
        if (fundamentals == null) return _buildFundamentalTableMock();

        final List<(String, String, String)> data = [
          ('PER', fundamentals.per?.toStringAsFixed(1) ?? '-', _scoreLabel(fundamentals.per, 0, 20)),
          ('ROE', fundamentals.roe != null ? '${fundamentals.roe!.toStringAsFixed(1)}%' : '-', _scoreLabel(fundamentals.roe, 15, 30)),
          ('Dividend Yield', fundamentals.dividendYield != null ? '${fundamentals.dividendYield!.toStringAsFixed(1)}%' : '-', _scoreLabel(fundamentals.dividendYield, 3, 8)),
          ('EPS', fundamentals.eps?.toStringAsFixed(0) ?? '-', '-'),
          ('Dette/Capitaux', fundamentals.debtToEquity?.toStringAsFixed(1) ?? '-', _scoreLabelReverse(fundamentals.debtToEquity, 0, 1)),
          ('Croissance CA', fundamentals.revenueGrowth != null ? '${fundamentals.revenueGrowth!.toStringAsFixed(1)}%' : '-', _scoreLabel(fundamentals.revenueGrowth, 5, 20)),
        ];

        return _buildFundamentalCard(data);
      },
    );
  }

  Widget _buildFundamentalTableMock() {
    final data = [
      ('PER', '12.5', '80/100'),
      ('ROE', '28.5%', '95/100'),
      ('Dividend Yield', '6.5%', '85/100'),
      ('EPS', '2 080 FCFA', '-'),
      ('Dette/Capitaux', '0.4', '90/100'),
      ('Croissance CA', '15.0%', '70/100'),
    ];
    return _buildFundamentalCard(data);
  }

  Widget _buildFundamentalCard(List<(String, String, String)> data) {
    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Column(
        children: data.map((d) {
          final (label, value, score) = d;
          return ListTile(
            leading: Text(label,
                style: const TextStyle(fontWeight: FontWeight.bold)),
            trailing: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(value,
                    style: const TextStyle(fontWeight: FontWeight.bold)),
                const SizedBox(width: 16),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.green.shade100,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    score,
                    style: TextStyle(
                      color: Colors.green.shade800,
                      fontWeight: FontWeight.bold,
                      fontSize: 12,
                    ),
                  ),
                ),
              ],
            ),
          );
        }).toList(),
      ),
    );
  }

  Widget _buildActionButtons() {
    return Row(
      children: [
        Expanded(
          child: ElevatedButton(
            onPressed: () {},
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.green,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(vertical: 16),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text(
              'ACHETER',
              style: TextStyle(fontWeight: FontWeight.bold),
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: ElevatedButton(
            onPressed: () {},
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(vertical: 16),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text(
              'VENDRE',
              style: TextStyle(fontWeight: FontWeight.bold),
            ),
          ),
        ),
      ],
    );
  }

  String _scoreLabel(double? value, double low, double high) {
    if (value == null) return '-';
    if (value >= high) return '90/100';
    if (value >= (low + high) / 2) return '75/100';
    if (value >= low) return '60/100';
    return '40/100';
  }

  String _scoreLabelReverse(double? value, double low, double high) {
    if (value == null) return '-';
    if (value <= low) return '95/100';
    if (value <= high) return '80/100';
    return '50/100';
  }
}
