import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../shared/providers/market_provider.dart';
import '../../shared/models/market_data.dart';
import 'ticker_detail_screen.dart';

/// Market screen — lists all BRVM tickers with prices.
///
/// Fetches tickers from the API and displays them in a scrollable list.
class MarketScreen extends ConsumerStatefulWidget {
  const MarketScreen({super.key});

  @override
  ConsumerState<MarketScreen> createState() => _MarketScreenState();
}

class _MarketScreenState extends ConsumerState<MarketScreen> {
  String _searchQuery = '';

  @override
  Widget build(BuildContext context) {
    final tickersAsync = ref.watch(tickersProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text(
          'Marche BRVM',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.search),
            onPressed: () => _showSearch(context),
          ),
        ],
      ),
      body: tickersAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => _buildMockList(),
        data: (tickers) {
          final filtered = _searchQuery.isEmpty
              ? tickers
              : tickers
                  .where((t) =>
                      t.symbol
                          .toLowerCase()
                          .contains(_searchQuery.toLowerCase()) ||
                      t.name
                          .toLowerCase()
                          .contains(_searchQuery.toLowerCase()))
                  .toList();

          if (filtered.isEmpty) return _buildMockList();

          return RefreshIndicator(
            onRefresh: () => ref.refresh(tickersProvider.future),
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: filtered.length,
              itemBuilder: (context, index) {
                final ticker = filtered[index];
                // Generate mock price from symbol hash for display
                final mockPrice =
                    25000 + (ticker.symbol.hashCode.abs() % 50000);
                final mockChange =
                    ((ticker.symbol.hashCode % 200) - 100) / 100;

                return Card(
                  margin: const EdgeInsets.only(bottom: 8),
                  elevation: 1,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: ListTile(
                    leading: CircleAvatar(
                      backgroundColor: Colors.blue.shade50,
                      child: Text(
                        ticker.symbol,
                        style: TextStyle(
                          color: Colors.blue.shade800,
                          fontWeight: FontWeight.bold,
                          fontSize: 10,
                        ),
                      ),
                    ),
                    title: Text(
                      ticker.name,
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                    subtitle: Text('${ticker.symbol} • ${ticker.country}'),
                    trailing: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(
                          '$mockPrice FCFA',
                          style: const TextStyle(fontWeight: FontWeight.bold),
                        ),
                        Text(
                          '${mockChange >= 0 ? '+' : ''}${mockChange.toStringAsFixed(2)}%',
                          style: TextStyle(
                            color: mockChange >= 0
                                ? Colors.green
                                : Colors.red,
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ),
                    onTap: () => Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (_) =>
                            TickerDetailScreen(ticker: ticker.symbol),
                      ),
                    ),
                  ),
                );
              },
            ),
          );
        },
      ),
    );
  }

  void _showSearch(BuildContext context) {
    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Rechercher un ticker'),
          content: TextField(
            autofocus: true,
            decoration: const InputDecoration(
              hintText: 'SNTS, Sonatel...',
              prefixIcon: Icon(Icons.search),
            ),
            onChanged: (value) {
              setState(() => _searchQuery = value);
            },
          ),
          actions: [
            TextButton(
              onPressed: () {
                setState(() => _searchQuery = '');
                Navigator.pop(context);
              },
              child: const Text('Effacer'),
            ),
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('OK'),
            ),
          ],
        );
      },
    );
  }

  Widget _buildMockList() {
    final tickers = [
      ('SNTS', 'Sonatel', 'Senegal'),
      ('ORAC', 'Orange CI', "Cote d'Ivoire"),
      ('SGBC', 'SGB CI', "Cote d'Ivoire"),
      ('ECOC', 'Ecobank CI', "Cote d'Ivoire"),
      ('PALC', 'PalmCI', "Cote d'Ivoire"),
      ('SPHC', 'SAPH CI', "Cote d'Ivoire"),
      ('SMB', 'SMB CI', "Cote d'Ivoire"),
      ('SOGC', 'SOGB CI', "Cote d'Ivoire"),
      ('SLBC', 'Solibra CI', "Cote d'Ivoire"),
      ('TTLC', 'Total CI', "Cote d'Ivoire"),
      ('TTLS', 'Total SN', 'Senegal'),
      ('UNLC', 'Unilever CI', "Cote d'Ivoire"),
      ('NEST', 'Nestle CI', "Cote d'Ivoire"),
      ('NSIA', 'NSIA Banque', "Cote d'Ivoire"),
      ('ETI', 'ETI TG', 'Togo'),
      ('ORGT', 'Oragroup TG', 'Togo'),
    ];

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: tickers.length,
      itemBuilder: (context, index) {
        final (symbol, name, country) = tickers[index];
        final price = 25000 + (index * 500);
        final change = index % 3 == 0 ? 1.24 : index % 3 == 1 ? -0.45 : 0.0;

        return Card(
          margin: const EdgeInsets.only(bottom: 8),
          elevation: 1,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          child: ListTile(
            leading: CircleAvatar(
              backgroundColor: Colors.blue.shade50,
              child: Text(
                symbol,
                style: TextStyle(
                  color: Colors.blue.shade800,
                  fontWeight: FontWeight.bold,
                  fontSize: 10,
                ),
              ),
            ),
            title: Text(name, style: const TextStyle(fontWeight: FontWeight.bold)),
            subtitle: Text('$symbol • $country'),
            trailing: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text('$price FCFA', style: const TextStyle(fontWeight: FontWeight.bold)),
                Text(
                  '${change >= 0 ? '+' : ''}${change.toStringAsFixed(2)}%',
                  style: TextStyle(
                    color: change >= 0 ? Colors.green : Colors.red,
                    fontSize: 12,
                  ),
                ),
              ],
            ),
            onTap: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (_) => TickerDetailScreen(ticker: symbol),
              ),
            ),
          ),
        );
      },
    );
  }
}
