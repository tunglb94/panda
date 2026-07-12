/// Dynamic-pricing signal for the current trip.
///
/// `backend/services/pricing` now has a real Dynamic Pricing Engine
/// (`PricingPipeline`/`PricingRule`s for Demand Surge, Night, Holiday, Rain,
/// Peak Hour, Airport) — see the Pricing Service CHANGELOG entry — but every
/// rule ships disabled by default and, more importantly, `FareBreakdown` in
/// `pricing.proto` has no surge-multiplier or surge-label field at all. The
/// app has no way to know a trip is surged even when the engine is later
/// turned on, until that proto is extended. [SurgeIndicator] is fully built
/// against this model and stays dormant (never constructed with a non-null
/// value anywhere in the app) until that field exists.
class SurgeInfo {
  const SurgeInfo({required this.label, required this.explanation});

  /// Short rider-facing label, e.g. "Giá đang thay đổi".
  final String label;

  /// Plain-language explanation shown on tap — per BRB §2.13.1, framed as a
  /// temporary supply/demand signal, never as a threat or a penalty.
  final String explanation;
}
