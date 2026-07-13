import 'package:flutter/material.dart';

/// Admin Web's single [ThemeData] — kept intentionally plain (Phần 12: "không
/// cần framework Admin, không animation, ưu tiên tốc độ"). Same brand green
/// as apps/driver and apps/rider (Color(0xFF1A8C4E)), everything else is
/// Material 3 defaults.
abstract final class AppTheme {
  static const primary = Color(0xFF1A8C4E);
  static const error = Color(0xFFDC2626);

  static ThemeData get light {
    final colorScheme = ColorScheme.fromSeed(
      seedColor: primary,
      brightness: Brightness.light,
      primary: primary,
      error: error,
    );
    const border = OutlineInputBorder();
    return ThemeData(
      useMaterial3: true,
      colorScheme: colorScheme,
      appBarTheme: const AppBarTheme(elevation: 0, centerTitle: false, scrolledUnderElevation: 0),
      cardTheme: CardThemeData(
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
          side: BorderSide(color: colorScheme.outlineVariant),
        ),
      ),
      dataTableTheme: const DataTableThemeData(headingRowHeight: 44, dataRowMinHeight: 56, dataRowMaxHeight: 64),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(backgroundColor: primary, foregroundColor: Colors.white),
      ),
      inputDecorationTheme: const InputDecorationTheme(isDense: true, filled: true, border: border, enabledBorder: border),
    );
  }
}
