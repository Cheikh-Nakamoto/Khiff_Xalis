import 'package:flutter/material.dart';

/// Horizontal scrolling BRVM indices overview.
///
/// Displays Composite, BRVM 30, Prestige, and Finance indices
/// with current value and daily change.
class IndicesHorizontal extends StatelessWidget {
  const IndicesHorizontal({super.key});

  @override
  Widget build(BuildContext context) {
    // BRVM indices — these could come from a dedicated API endpoint.
    // For now they are static as the Go API does not expose an index endpoint yet.
    final indices = [
      _IndexData('BRVM Composite', '249.85', '+1.24%', true),
      _IndexData('BRVM 30', '312.40', '+0.89%', true),
      _IndexData('BRVM Prestige', '156.20', '-0.45%', false),
      _IndexData('BRVM Finance', '198.75', '+2.10%', true),
    ];

    return SizedBox(
      height: 120,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        itemCount: indices.length,
        separatorBuilder: (_, __) => const SizedBox(width: 12),
        itemBuilder: (context, index) {
          final idx = indices[index];
          return Card(
            elevation: 2,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
            child: Container(
              width: 160,
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    idx.name,
                    style: const TextStyle(fontSize: 13, color: Colors.grey),
                  ),
                  Text(
                    idx.value,
                    style: const TextStyle(
                      fontSize: 22,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Text(
                    idx.change,
                    style: TextStyle(
                      color: idx.isPositive ? Colors.green : Colors.red,
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}

class _IndexData {
  final String name;
  final String value;
  final String change;
  final bool isPositive;

  _IndexData(this.name, this.value, this.change, this.isPositive);
}
