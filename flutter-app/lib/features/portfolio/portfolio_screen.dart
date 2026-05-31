import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../shared/providers/portfolio_provider.dart';
import '../../shared/models/portfolio_position.dart';

/// Portfolio screen — shows user's holdings and P&L.
///
/// Fetches from the API via [portfolioProvider] with mock fallback.
class PortfolioScreen extends ConsumerWidget {
  const PortfolioScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final portfolioAsync = ref.watch(portfolioProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text(
          'Mon Portefeuille',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.read(portfolioProvider.notifier).refresh(),
          ),
        ],
      ),
      body: portfolioAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => _buildPortfolioBody(context, _mockSummary),
        data: (portfolio) =>
            _buildPortfolioBody(context, portfolio),
      ),
    );
  }

  Widget _buildPortfolioBody(
      BuildContext context, PortfolioSummary portfolio) {
    final pnlSign = portfolio.totalPnL >= 0 ? '+' : '';

    return Column(
      children: [
        // Header
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [Colors.blue.shade700, Colors.blue.shade900],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
          ),
          child: Column(
            children: [
              const Text(
                'Valeur Totale',
                style: TextStyle(color: Colors.white70),
              ),
              Text(
                _formatCurrency(portfolio.totalValue),
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 36,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                decoration: BoxDecoration(
                  color: portfolio.totalPnL >= 0
                      ? Colors.green.shade400
                      : Colors.red.shade400,
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  '$pnlSign${_formatCurrency(portfolio.totalPnL)} '
                  '($pnlSign${portfolio.totalPnLPct.toStringAsFixed(1)}%)',
                  style: const TextStyle(color: Colors.white),
                ),
              ),
            ],
          ),
        ),
        // Positions
        Expanded(
          child: portfolio.positions.isEmpty
              ? const Center(child: Text('Aucune position'))
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: portfolio.positions.length,
                  itemBuilder: (context, index) {
                    return _buildPositionTile(portfolio.positions[index]);
                  },
                ),
        ),
      ],
    );
  }

  Widget _buildPositionTile(PortfolioPosition p) {
    final pnlSign = p.unrealizedPnL >= 0 ? '+' : '';

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        leading: CircleAvatar(child: Text(p.ticker)),
        title: Text(
          p.name.isNotEmpty ? p.name : p.ticker,
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Text('${p.quantity} actions @ ${p.avgBuyPrice.toStringAsFixed(0)} FCFA'),
        trailing: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text(
              '${_formatCurrency(p.marketValue)} FCFA',
              style: const TextStyle(fontWeight: FontWeight.bold),
            ),
            Text(
              '$pnlSign${_formatCurrency(p.unrealizedPnL)} FCFA',
              style: TextStyle(
                color: p.isProfit ? Colors.green : Colors.red,
                fontSize: 12,
              ),
            ),
          ],
        ),
      ),
    );
  }

  static String _formatCurrency(double value) {
    final parts = value.toStringAsFixed(0).split('');
    final result = <String>[];
    int count = 0;
    for (int i = parts.length - 1; i >= 0; i--) {
      if (count > 0 && count % 3 == 0) {
        result.add(' ');
      }
      result.add(parts[i]);
      count++;
    }
    return result.reversed.join();
  }
}

const _mockSummary = PortfolioSummary(
  totalValue: 3975000,
  totalPnL: 125000,
  positions: [
    PortfolioPosition(
      ticker: 'SNTS', name: 'Sonatel', quantity: 100,
      avgBuyPrice: 25000, currentPrice: 26000,
      marketValue: 2600000, unrealizedPnL: 100000, unrealizedPnLPct: 4.0,
    ),
    PortfolioPosition(
      ticker: 'SGBC', name: 'SGB CI', quantity: 50,
      avgBuyPrice: 27000, currentPrice: 27500,
      marketValue: 1375000, unrealizedPnL: 25000, unrealizedPnLPct: 1.85,
    ),
    PortfolioPosition(
      ticker: 'ORAC', name: 'Orange CI', quantity: 200,
      avgBuyPrice: 14500, currentPrice: 14650,
      marketValue: 2930000, unrealizedPnL: 30000, unrealizedPnLPct: 1.03,
    ),
    PortfolioPosition(
      ticker: 'PALC', name: 'PalmCI', quantity: 300,
      avgBuyPrice: 9600, currentPrice: 9515,
      marketValue: 2854500, unrealizedPnL: -25500, unrealizedPnLPct: -0.89,
    ),
    PortfolioPosition(
      ticker: 'SPHC', name: 'SAPH CI', quantity: 150,
      avgBuyPrice: 8000, currentPrice: 7880,
      marketValue: 1182000, unrealizedPnL: -18000, unrealizedPnLPct: -1.50,
    ),
  ],
);
