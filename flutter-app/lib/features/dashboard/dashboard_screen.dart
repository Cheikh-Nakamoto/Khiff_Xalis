import 'package:flutter/material.dart';
import 'widgets/portfolio_card.dart';
import 'widgets/indices_horizontal.dart';
import 'widgets/signals_list.dart';
import '../market/widgets/heatmap_widget.dart';

/// Dashboard screen — the main "home" tab.
///
/// Shows portfolio summary, market indices, top signals, and sector heatmap.
class DashboardScreen extends StatelessWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text(
          'BRVM Trading Engine',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.notifications_outlined),
            onPressed: () {},
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Portfolio Summary Card
            const PortfolioCard(),
            const SizedBox(height: 20),

            // Market Overview (indices)
            Text(
              'Vue du Marche',
              style: Theme.of(context)
                  .textTheme
                  .titleLarge
                  ?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            const IndicesHorizontal(),
            const SizedBox(height: 20),

            // Top Signals
            Text(
              'Top Signaux du Jour',
              style: Theme.of(context)
                  .textTheme
                  .titleLarge
                  ?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            const SignalsList(),
            const SizedBox(height: 20),

            // Sector Heatmap
            Text(
              'Heatmap Sectorielle',
              style: Theme.of(context)
                  .textTheme
                  .titleLarge
                  ?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            const HeatmapWidget(),
          ],
        ),
      ),
    );
  }
}
