import 'package:flutter/material.dart';

/// Central color palette for Panda — kept in lockstep with `apps/driver`'s
/// `AppColors` (same values, same names) so the two apps read as one
/// product family. Brand seed is `0xFF1A8C4E`, unchanged from the rider
/// app's original `AppTheme` — this file just promotes what used to be
/// three private constants buried in `app_theme.dart` into the same named,
/// reusable token set driver already has, plus the semantic/surface colors
/// driver's screens depend on that rider never had (error/warning/info,
/// textTertiary, border, divider, primaryDark).
///
/// Every color used more than once anywhere in the app should have a named
/// constant here rather than being written inline as `Colors.grey.shade200`
/// or a raw `Color(0xFF...)` literal at the call site.
abstract final class AppColors {
  // ─── Brand ──────────────────────────────────────────────────────────────
  static const Color primary = Color(0xFF1A8C4E);
  static const Color primaryDark = Color(0xFF13703C);
  static const Color primaryLight = Color(0xFFE8F5ED);

  // ─── Semantic ───────────────────────────────────────────────────────────
  /// Brand green doubles as the "success" color — this is a green-branded
  /// app, so a distinct success hue would fight the brand rather than
  /// reinforce it. Use [primary] for success states.
  static const Color warning = Color(0xFFF59E0B);
  static const Color error = Color(0xFFDC2626);
  static const Color errorLight = Color(0xFFFCA5A5);
  static const Color info = Color(0xFF2563EB);

  // ─── Text ───────────────────────────────────────────────────────────────
  static const Color textPrimary = Color(0xFF1C1C1E);
  static const Color textSecondary = Color(0xFF6B7280);
  static const Color textTertiary = Color(0xFF9CA3AF);
  static const Color textOnPrimary = Colors.white;

  // ─── Surfaces ───────────────────────────────────────────────────────────
  static const Color surface = Colors.white;
  static const Color surfaceAlt = Color(0xFFF8F9FA);
  static const Color border = Color(0xFFE5E7EB);
  static const Color divider = Color(0xFFF3F4F6);

  // ─── Overlays ───────────────────────────────────────────────────────────
  static Color scrim(double opacity) => Colors.black.withValues(alpha: opacity);
  static Color tint(Color color, double opacity) => color.withValues(alpha: opacity);
}
