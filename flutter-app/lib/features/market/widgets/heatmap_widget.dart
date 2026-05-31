import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../shared/providers/market_provider.dart';

/// Sector heatmap widget displayed on the dashboard.
///
/// Shows color-coded sector performance using real data from the API,
/// falling back to mock data when unavailable.
class HeatmapWidget extends ConsumerWidget {
  const HeatmapWidget({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final sectorsAsync = ref.watch(sectorHeatmapProvider);

    return sectorsAsync.when(
      loading: () => const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: CircularProgressIndicator(),
        ),
      ),
      error: (_, __) => _buildMockHeatmap(context),
      data: (sectors) {
        if (sectors.isEmpty) return _buildMockHeatmap(context);

        final totalTickers =
            sectors.fold<int>(0, (sum, s) => sum + s.tickerCount);
        if (totalTickers == 0) return _buildMockHeatmap(context);

        return Wrap(
          spacing: 8,
          runSpacing: 8,
          children: sectors.map((s) {
            final sizeFraction = s.tickerCount / totalTickers;
            final color = _colorForChange(s.changePercent);
            final width =
                (MediaQuery.of(context).size.width - 48) * sizeFraction;

            return Container(
              width: width.clamp(80, double.infinity),
              height: 60,
              decoration: BoxDecoration(
                color: color,
                borderRadius: BorderRadius.circular(8),
              ),
              padding: const EdgeInsets.all(8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    s.sector,
                    style: const TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Text(
                    '${s.changePercent >= 0 ? '+' : ''}${s.changePercent.toStringAsFixed(1)}%',
                    style: const TextStyle(fontSize: 12),
                  ),
                ],
              ),
            );
          }).toList(),
        );
      },
    );
  }

  Widget _buildMockHeatmap(BuildContext context) {
    final sectors = [
      _MockSector('Telecom', '+2.4%', 0.22, Colors.green.shade400),
      _MockSector('Finance', '+1.8%', 0.21, Colors.green.shade300),
      _MockSector('Industrie', '+0.5%', 0.20, Colors.green.shade100),
      _MockSector('Agriculture', '+3.1%', 0.10, Colors.green.shade500),
      _MockSector('Consommation', '-0.2%', 0.11, Colors.red.shade100),
      _MockSector('Distribution', '+0.8%', 0.09, Colors.green.shade200),
      _MockSector('Services', '+0.1%', 0.07, Colors.grey.shade200),
    ];

    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: sectors.map((s) {
        return Container(
          width: (MediaQuery.of(context).size.width - 48) * s.size,
          height: 60,
          decoration: BoxDecoration(
            color: s.color,
            borderRadius: BorderRadius.circular(8),
          ),
          padding: const EdgeInsets.all(8),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                s.name,
                style: const TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                ),
              ),
              Text(s.change, style: const TextStyle(fontSize: 12)),
            ],
          ),
        );
      }).toList(),
    );
  }

  Color _colorForChange(double change) {
    if (change > 2) return Colors.green.shade500;
    if (change > 1) return Colors.green.shade400;
    if (change > 0.5) return Colors.green.shade300;
    if (change > 0) return Colors.green.shade100;
    if (change > -0.5) return Colors.red.shade100;
    return Colors.red.shade200;
  }
}

class _MockSector {
  final String name;
  final String change;
  final double size;
  final Color color;

  _MockSector(this.name, this.change, this.size, this.color);
}
