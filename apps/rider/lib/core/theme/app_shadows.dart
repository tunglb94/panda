import 'package:flutter/material.dart';

/// Standard elevation shadows. Mirrors `apps/driver`'s `AppShadows`. Panda
/// uses Material 3 with `elevation: 0` everywhere (see `AppTheme`) and
/// expresses depth through soft, deliberate `BoxShadow`s instead of the
/// default Material elevation tint.
abstract final class AppShadows {
  /// Standard resting card shadow — used by nearly every card-like surface.
  static List<BoxShadow> card = [
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.04),
      blurRadius: 10,
      offset: const Offset(0, 2),
    ),
  ];

  /// Stronger shadow for floating/emphasized elements (booking sheet, driver
  /// card, dialogs rendered outside the default Material dialog surface).
  static List<BoxShadow> raised = [
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.10),
      blurRadius: 20,
      offset: const Offset(0, 6),
    ),
  ];

  /// Colored glow for a tinted emphasis surface (e.g. a primary-colored
  /// circular badge). Pass the surface's own color in.
  static List<BoxShadow> glow(Color color) => [
        BoxShadow(
          color: color.withValues(alpha: 0.35),
          blurRadius: 12,
          offset: const Offset(0, 4),
        ),
      ];

  static const List<BoxShadow> none = [];
}
