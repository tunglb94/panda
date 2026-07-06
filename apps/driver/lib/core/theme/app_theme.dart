import 'package:flutter/material.dart';

abstract final class AppTheme {
  static const _seed = Color(0xFF1A8C4E);

  static ThemeData get light => ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(seedColor: _seed),
        scaffoldBackgroundColor: const Color(0xFFF8F9FA),
        navigationBarTheme: const NavigationBarThemeData(
          indicatorColor: Color(0xFFE8F5ED),
        ),
      );
}
