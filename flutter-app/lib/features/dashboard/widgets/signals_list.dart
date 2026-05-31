import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../shared/providers/signal_provider.dart';
import '../../../shared/models/signal.dart';
import '../../market/ticker_detail_screen.dart';

/// Top signals list widget for the dashboard.
///
/// Displays the top 5 trading signals with scores and action badges.
class SignalsList extends ConsumerWidget {
  const SignalsList({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final signalsAsync = ref.watch(topSignalsProvider);

    return signalsAsync.when(
      loading: () => const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: CircularProgressIndicator(),
        ),
      ),
      error: (_, __) => _buildMockList(context),
      data: (signals) {
        if (signals.isEmpty) return _buildMockList(context);
        return _buildSignalsList(context, signals);
      },
    );
  }

  Widget _buildSignalsList(BuildContext context, List<Signal> signals) {
    return Column(
      children: signals.map((s) => _buildSignalTile(context, s)).toList(),
    );
  }

  Widget _buildSignalTile(BuildContext context, Signal s) {
    final color = s.isBuy
        ? Colors.green
        : s.isSell
            ? Colors.red
            : Colors.orange;

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      elevation: 1,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: color.withOpacity(0.2),
          child: Text(
            s.tickerSymbol,
            style: TextStyle(
              color: color,
              fontWeight: FontWeight.bold,
              fontSize: 11,
            ),
          ),
        ),
        title: Text(
          s.tickerName.isNotEmpty ? s.tickerName : s.tickerSymbol,
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Text('Score: ${s.compositeScore.toStringAsFixed(1)} | ${s.labelFr}'),
        trailing: Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          decoration: BoxDecoration(
            color: color.withOpacity(0.1),
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: color),
          ),
          child: Text(
            s.labelFr,
            style: TextStyle(
              color: color,
              fontWeight: FontWeight.bold,
              fontSize: 12,
            ),
          ),
        ),
        onTap: () => Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => TickerDetailScreen(ticker: s.tickerSymbol),
          ),
        ),
      ),
    );
  }

  Widget _buildMockList(BuildContext context) {
    final mockSignals = [
      _MockSignal('SIBC', 'SIB CI', 'ACHAT FORT', 88.8, Colors.green),
      _MockSignal('SNTS', 'Sonatel', 'ACHAT', 87.4, Colors.green),
      _MockSignal('SPHC', 'SAPH CI', 'ACHAT FORT', 84.1, Colors.green),
      _MockSignal('ETI', 'ETI TG', 'ACHAT', 65.0, Colors.orange),
      _MockSignal('UNXC', 'Uniwax', 'VENTE', 38.9, Colors.red),
    ];

    return Column(
      children: mockSignals.map((s) {
        return Card(
          margin: const EdgeInsets.only(bottom: 8),
          elevation: 1,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          child: ListTile(
            leading: CircleAvatar(
              backgroundColor: s.color.withOpacity(0.2),
              child: Text(
                s.ticker,
                style: TextStyle(
                  color: s.color,
                  fontWeight: FontWeight.bold,
                  fontSize: 11,
                ),
              ),
            ),
            title: Text(s.name, style: const TextStyle(fontWeight: FontWeight.bold)),
            subtitle: Text('Score: ${s.score} | ${s.signal}'),
            trailing: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(
                color: s.color.withOpacity(0.1),
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: s.color),
              ),
              child: Text(
                s.signal,
                style: TextStyle(
                  color: s.color,
                  fontWeight: FontWeight.bold,
                  fontSize: 12,
                ),
              ),
            ),
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (_) => TickerDetailScreen(ticker: s.ticker),
              ),
            ),
          ),
        );
      }).toList(),
    );
  }
}

class _MockSignal {
  final String ticker;
  final String name;
  final String signal;
  final double score;
  final Color color;

  _MockSignal(this.ticker, this.name, this.signal, this.score, this.color);
}
