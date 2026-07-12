/// The one spacing scale for Panda: 4 · 8 · 12 · 16 · 20 · 24 · 32.
/// Mirrors `apps/driver`'s `AppSpacing` exactly.
///
/// Every `EdgeInsets`/`SizedBox`/gap value in the app should resolve to one
/// of these — no bespoke `10`, `14`, `18`, `22` one-offs. If a spot genuinely
/// needs something off-scale, that's a signal to reconsider the layout
/// before reaching for a magic number.
abstract final class AppSpacing {
  static const double xs = 4;
  static const double sm = 8;
  static const double md = 12;
  static const double lg = 16;
  static const double xl = 20;
  static const double xxl = 24;
  static const double xxxl = 32;
}
