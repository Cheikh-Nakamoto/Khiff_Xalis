import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../shared/providers/signal_provider.dart';
import '../../shared/models/signal.dart';
import '../market/ticker_detail_screen.dart';

/// Signals screen — displays all trading signals with filter tabs.
///
/// Fetches signals from the API via [signalsProvider] with tab-based filtering.
class SignalsScreen extends ConsumerStatefulWidget {
  const SignalsScreen({super.key});

  @override
  ConsumerState<SignalsScreen> createState() => _SignalsScreenState();
}

class _SignalsScreenState extends ConsumerState<SignalsScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabController;
  static const _tabs = [
    ('ACHAT', 'BUY'),
    ('NEUTRE', 'HOLD'),
    ('VENTE', 'SELL'),
  ];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: _tabs.length, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text(
          'Signaux de Trading',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(text: 'ACHAT', icon: Icon(Icons.arrow_upward)),
            Tab(text: 'NEUTRE', icon: Icon(Icons.remove)),
            Tab(text: 'VENTE', icon: Icon(Icons.arrow_downward)),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _SignalList(signalType: 'BUY'),
          _SignalList(signalType: 'HOLD'),
          _SignalList(signalType: 'SELL'),
        ],
      ),
    );
  }
}

/// Signal list for a specific signal type.
class _SignalList extends ConsumerWidget {
  final String signalType;
  const _SignalList({required this.signalType});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final signalsAsync = ref.watch(signalsProvider(signalType));

    return signalsAsync.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (_, __) => _buildMockList(context),
      data: (signals) {
        if (signals.isEmpty) return _buildMockList(context);

        return RefreshIndicator(
          onRefresh: () => ref.refresh(signalsProvider(signalType).future),
          child: ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: signals.length,
            itemBuilder: (context, index) {
              return _buildSignalCard(context, signals[index]);
            },
          ),
        );
      },
    );
  }

  Widget _buildSignalCard(BuildContext context, Signal s) {
    final color = s.isBuy
        ? Colors.green
        : s.isSell
            ? Colors.red
            : Colors.orange;

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    CircleAvatar(
                      backgroundColor: color.withOpacity(0.2),
                      child: Text(
                        s.tickerSymbol,
                        style: TextStyle(
                          color: color,
                          fontWeight: FontWeight.bold,
                          fontSize: 10,
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          s.tickerName.isNotEmpty
                              ? s.tickerName
                              : s.tickerSymbol,
                          style: const TextStyle(
                            fontWeight: FontWeight.bold,
                            fontSize: 16,
                          ),
                        ),
                        Text(
                          'Score: ${s.compositeScore.toStringAsFixed(1)}',
                          style: TextStyle(color: Colors.grey.shade600),
                        ),
                      ],
                    ),
                  ],
                ),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  decoration: BoxDecoration(
                    color: color,
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Text(
                    s.labelFr,
                    style: const TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                      fontSize: 12,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildMetric('Confiance',
                    '${(s.confidence * 100).toStringAsFixed(0)}%', color),
                _buildMetric('Score',
                    s.compositeScore.toStringAsFixed(1), color),
                _buildMetric(
                    'Signaux', '${s.reasons.length}', Colors.blue),
              ],
            ),
            if (s.reasons.isNotEmpty) ...[
              const SizedBox(height: 12),
              Wrap(
                spacing: 6,
                runSpacing: 4,
                children: s.reasons
                    .take(3)
                    .map((r) => Chip(
                          label: Text(r, style: const TextStyle(fontSize: 11)),
                          materialTapTargetSize:
                              MaterialTapTargetSize.shrinkWrap,
                          visualDensity: VisualDensity.compact,
                        ))
                    .toList(),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildMetric(String label, String value, Color color) {
    return Column(
      children: [
        Text(
          value,
          style: TextStyle(
            fontSize: 18,
            fontWeight: FontWeight.bold,
            color: color,
          ),
        ),
        Text(
          label,
          style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
        ),
      ],
    );
  }

  Widget _buildMockList(BuildContext context) {
    final mockSignals = signalType == 'BUY'
        ? [
            _MockSignal('SIBC', 'SIB CI', 'STRONG_BUY', 88.8, 0.92,
                ['RSI survente', 'PER sous-evalue']),
            _MockSignal('SNTS', 'Sonatel SN', 'BUY', 87.4, 0.88,
                ['MACD haussier', 'Volume hausse']),
            _MockSignal('SPHC', 'SAPH CI', 'STRONG_BUY', 84.1, 0.85,
                ['Cassure resistance']),
            _MockSignal('PALC', 'PalmCI', 'BUY', 65.7, 0.70,
                ['Secteur en hausse']),
          ]
        : signalType == 'SELL'
            ? [
                _MockSignal('UNXC', 'Uniwax CI', 'SELL', 38.9, 0.72,
                    ['RSI surachat', 'Dette elevee']),
              ]
            : [
                _MockSignal('ETI', 'ETI TG', 'HOLD', 55.0, 0.55,
                    ['Consolidation']),
              ];

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: mockSignals.length,
      itemBuilder: (context, index) {
        return _buildSignalCard(context, mockSignals[index].toSignal());
      },
    );
  }
}

class _MockSignal {
  final String ticker;
  final String name;
  final String signal;
  final double score;
  final double confidence;
  final List<String> reasons;

  _MockSignal(
      this.ticker, this.name, this.signal, this.score, this.confidence, this.reasons);

  Signal toSignal() => Signal(
        tickerSymbol: ticker,
        tickerName: name,
        compositeScore: score,
        signal: signal,
        confidence: confidence,
        reasons: reasons,
      );
}
