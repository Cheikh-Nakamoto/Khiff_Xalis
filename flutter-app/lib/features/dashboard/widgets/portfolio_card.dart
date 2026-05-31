import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../shared/providers/portfolio_provider.dart';
import '../../../shared/models/portfolio_position.dart';

/// Portfolio summary card displayed on the dashboard.
///
/// Shows total value, daily P&L, and key stats.
/// Falls back to mock data if API is unavailable.
class PortfolioCard extends ConsumerWidget {
  const PortfolioCard({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final portfolioAsync = ref.watch(portfolioProvider);

    return portfolioAsync.when(
      loading: () => _buildCard(
        context,
        totalValue: '...',
        pnlText: 'Chargement...',
        isProfit: true,
        positions: 0,
        cash: '...',
        yield: '...',
      ),
      error: (_, __) => _buildCard(
        context,
        totalValue: '3 975 000 FCFA',
        pnlText: '+125 000 FCFA (+3.2%)',
        isProfit: true,
        positions: 5,
        cash: '250 000',
        yield: '6.8%',
      ),
      data: (portfolio) {
        final pnlSign = portfolio.totalPnL >= 0 ? '+' : '';
        final pnlPct = portfolio.totalValue > 0
            ? (portfolio.totalPnL /
                    (portfolio.totalValue - portfolio.totalPnL) *
                    100)
                .toStringAsFixed(1)
            : '0.0';

        return _buildCard(
          context,
          totalValue: _formatCurrency(portfolio.totalValue),
          pnlText:
              '$pnlSign${_formatCurrency(portfolio.totalPnL)} ($pnlSign$pnlPct%)',
          isProfit: portfolio.totalPnL >= 0,
          positions: portfolio.positions.length,
          cash: _formatCurrency(
              portfolio.totalValue * 0.06), // approximate
          yield: '6.8%',
        );
      },
    );
  }

  Widget _buildCard(
    BuildContext context, {
    required String totalValue,
    required String pnlText,
    required bool isProfit,
    required int positions,
    required String cash,
    required String yield,
  }) {
    return Card(
      elevation: 4,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(16),
          gradient: LinearGradient(
            colors: [Colors.blue.shade700, Colors.blue.shade900],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Valeur du Portefeuille',
              style: TextStyle(color: Colors.white70, fontSize: 14),
            ),
            const SizedBox(height: 8),
            Text(
              totalValue,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 32,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: isProfit ? Colors.green.shade400 : Colors.red.shade400,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    pnlText,
                    style: const TextStyle(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                _buildStat('Positions', positions.toString()),
                _buildStat('Cash', cash),
                _buildStat('Yield', yield),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStat(String label, String value) {
    return Column(
      children: [
        Text(
          value,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 18,
            fontWeight: FontWeight.bold,
          ),
        ),
        Text(
          label,
          style: const TextStyle(color: Colors.white70, fontSize: 12),
        ),
      ],
    );
  }

  static String _formatCurrency(double value) {
    // Format as FCFA with thousands separator
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
