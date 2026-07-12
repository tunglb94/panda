/// Icon size scale. Mirrors `apps/driver`'s `AppIconSize`. Pick the size
/// that matches the icon's role, not its visual container — e.g. an icon
/// inside a 52px avatar circle is still [lg] (24), the circle's own size is
/// a spacing/radius concern, not an icon-size one.
abstract final class AppIconSize {
  /// Inline with dense text (list trailing icons, chip icons).
  static const double sm = 16;

  /// Default icon size — buttons, list leading icons, app bar actions.
  static const double md = 20;

  /// Standalone/emphasized icons (card leading icon, status icon).
  static const double lg = 24;

  /// Hero icons inside a colored badge/avatar circle.
  static const double xl = 28;

  /// Empty-state / full-screen illustrative icons.
  static const double xxl = 40;
}
