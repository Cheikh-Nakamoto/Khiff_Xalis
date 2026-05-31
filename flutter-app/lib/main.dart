// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'core/theme/app_theme.dart';
import 'features/navigation/main_navigation.dart';

/// BRVM Trading Engine — Flutter application entry point.
///
/// Architecture:
///   lib/
///     core/config/env.dart          — Environment configuration
///     core/network/api_client.dart  — Dio HTTP client with auth interceptor
///     core/theme/app_theme.dart     — FlexColorScheme Material 3 themes
///     features/dashboard/           — Dashboard screen + widgets
///     features/market/              — Market list + ticker detail
///     features/portfolio/           — Portfolio screen
///     features/signals/             — Trading signals screen
///     features/settings/            — Settings screen
///     features/navigation/          — Bottom NavigationBar shell
///     shared/models/                — Data models (MarketData, Signal, etc.)
///     shared/providers/             — Riverpod providers (API + state)
///
/// Stack: Flutter + Riverpod + Dio + fl_chart + FlexColorScheme
/// API:   http://localhost:8080/api/v1 (Go Fiber backend)
void main() {
  runApp(const ProviderScope(child: BRVMTradingApp()));
}

class BRVMTradingApp extends StatelessWidget {
  const BRVMTradingApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'BRVM Trading Engine',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light,
      darkTheme: AppTheme.dark,
      themeMode: ThemeMode.system,
      home: const MainNavigationScreen(),
    );
  }
}
