import 'package:flex_color_scheme/flex_color_scheme.dart';
import 'package:flutter/material.dart';

/// Application theme using FlexColorScheme for consistent Material 3 styling.
class AppTheme {
  AppTheme._();

  static const _scheme = FlexScheme.blue;

  /// Light theme.
  static ThemeData get light {
    return FlexThemeData.light(
      scheme: _scheme,
      surfaceMode: FlexSurfaceMode.levelSurfacesLowScaffold,
      blendLevel: 7,
      subThemesData: const FlexSubThemesData(
        blendOnLevel: 10,
        blendOnColors: false,
        useTextTheme: true,
        useM2StyleDividerInM3: true,
        inputDecoratorBorderType: FlexInputBorderType.outline,
        inputDecoratorRadius: 12.0,
        chipRadius: 20.0,
        cardRadius: 12.0,
        dialogRadius: 16.0,
        fabRadius: 16.0,
        navigationBarIndicatorRadius: 12.0,
      ),
      visualDensity: FlexColorScheme.comfortablePlatformDensity,
      useMaterial3: true,
      fontFamily: 'Inter',
    );
  }

  /// Dark theme.
  static ThemeData get dark {
    return FlexThemeData.dark(
      scheme: _scheme,
      surfaceMode: FlexSurfaceMode.levelSurfacesLowScaffold,
      blendLevel: 13,
      subThemesData: const FlexSubThemesData(
        blendOnLevel: 20,
        useTextTheme: true,
        useM2StyleDividerInM3: true,
        inputDecoratorBorderType: FlexInputBorderType.outline,
        inputDecoratorRadius: 12.0,
        chipRadius: 20.0,
        cardRadius: 12.0,
        dialogRadius: 16.0,
        fabRadius: 16.0,
        navigationBarIndicatorRadius: 12.0,
      ),
      visualDensity: FlexColorScheme.comfortablePlatformDensity,
      useMaterial3: true,
      fontFamily: 'Inter',
    );
  }
}
