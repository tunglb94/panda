import 'package:flutter/material.dart';

/// Corner-radius scale. Small controls (chips, badges) use [sm]; the
/// standard card/button/input radius is [md]; hero cards and sheets use
/// [lg]/[xl]; [pill] is for fully-rounded shapes (status pills, avatars).
abstract final class AppRadius {
  static const double sm = 8;
  static const double md = 12;
  static const double lg = 16;
  static const double xl = 20;
  static const double pill = 999;

  static const BorderRadius smAll = BorderRadius.all(Radius.circular(sm));
  static const BorderRadius mdAll = BorderRadius.all(Radius.circular(md));
  static const BorderRadius lgAll = BorderRadius.all(Radius.circular(lg));
  static const BorderRadius xlAll = BorderRadius.all(Radius.circular(xl));
  static const BorderRadius pillAll = BorderRadius.all(Radius.circular(pill));

  /// Top-only rounding for bottom sheets and modal panels.
  static const BorderRadius topXl = BorderRadius.vertical(top: Radius.circular(xl));
}
