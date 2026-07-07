import 'package:flutter/material.dart';

class AppTheme {
  static const Color _primary = Color(0xFF1A8C4E);
  static const Color _secondaryText = Color(0xFF6B7280);
  static const Color _onSurface = Color(0xFF1C1C1E);

  static ThemeData get light => ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(
          seedColor: _primary,
          brightness: Brightness.light,
        ),
        appBarTheme: const AppBarTheme(
          backgroundColor: Colors.white,
          foregroundColor: _onSurface,
          elevation: 0,
          centerTitle: false,
          scrolledUnderElevation: 0,
        ),
        navigationBarTheme: NavigationBarThemeData(
          backgroundColor: Colors.white,
          indicatorColor: const Color(0xFFE8F5ED),
          iconTheme: WidgetStateProperty.resolveWith((states) {
            if (states.contains(WidgetState.selected)) {
              return const IconThemeData(color: _primary);
            }
            return IconThemeData(color: _secondaryText);
          }),
          labelTextStyle: WidgetStateProperty.resolveWith((states) {
            if (states.contains(WidgetState.selected)) {
              return const TextStyle(
                  color: _primary, fontWeight: FontWeight.w600, fontSize: 12);
            }
            return TextStyle(color: _secondaryText, fontSize: 12);
          }),
        ),
        elevatedButtonTheme: ElevatedButtonThemeData(
          style: ElevatedButton.styleFrom(
            backgroundColor: _primary,
            foregroundColor: Colors.white,
            minimumSize: const Size.fromHeight(52),
            shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12)),
            elevation: 0,
          ),
        ),
        outlinedButtonTheme: OutlinedButtonThemeData(
          style: OutlinedButton.styleFrom(
            foregroundColor: _primary,
            minimumSize: const Size.fromHeight(52),
            shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12)),
            side: const BorderSide(color: _primary),
          ),
        ),
        cardTheme: CardThemeData(
          elevation: 0,
          color: Colors.white,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: BorderSide(color: Colors.grey.shade200),
          ),
        ),
        scaffoldBackgroundColor: const Color(0xFFF8F9FA),
      );
}
