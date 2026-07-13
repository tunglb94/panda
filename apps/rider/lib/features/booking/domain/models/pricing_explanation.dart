import 'package:rider/shared/utils/currency_format.dart';

import 'fare_estimate.dart';
import 'surge_info.dart';
import 'voucher.dart';

/// One line in the "Tại sao giá này?" explanation sheet.
class PricingExplanationLine {
  const PricingExplanationLine(this.text, {this.isPositive = true});

  /// The explanation text, e.g. "8.3 km" or "Không áp dụng Surge".
  final String text;

  /// Whether this shows a green check (a fact that keeps the price as-is /
  /// lower) vs. an amber flag (a fact that raised the price, e.g. an active
  /// surcharge). Every line in the MVP is currently positive because Peak
  /// Hour/Night/Holiday/Rain/Demand Surge all ship disabled by default in
  /// production Pricing (see `PricingExplanation.build`'s doc comment) —
  /// the flag exists so a future surcharge line can render distinctly
  /// without a further redesign of this widget.
  final bool isPositive;
}

/// Builds the rider-facing "why this price" checklist — PHẦN 2 of the
/// Payment/Fare production-readiness pass. Explicitly rule-based, not AI:
/// every line is a deterministic fact computed from data already on this
/// screen (the fare estimate, the trip's distance/duration, the current
/// clock time, and whatever voucher/surge state was actually passed in) —
/// nothing here is generated or inferred.
///
/// The peak-hour check cites Business Rule Bible v1.0 §2.2.12 verbatim
/// (07:00–09:00 and 17:00–20:00, Monday–Friday, +10%) rather than a made-up
/// window. Whether that line reads "áp dụng"/"không áp dụng" is purely a
/// clock-time fact; it does NOT claim a peak-hour surcharge was actually
/// added to [fare], because Pricing's Dynamic Pricing Engine ships with
/// every rule (Peak Hour, Night, Holiday, Rain, Demand Surge) disabled by
/// default in production — see `backend/services/pricing/app/rules_defined.go`
/// `DefaultRuleConfigs()` and the Pricing V3 Engine CHANGELOG entry. The
/// checklist is written to stay true either way: it states a time-window
/// fact and separately confirms no surge was applied to this estimate.
abstract final class PricingExplanation {
  static List<PricingExplanationLine> build({
    required FareEstimate fare,
    required double distanceKm,
    required double durationMin,
    required DateTime requestTime,
    Voucher? voucher,
    SurgeInfo? surge,
  }) {
    final lines = <PricingExplanationLine>[
      PricingExplanationLine('Giá cơ bản ${formatMoney(fare.baseFare, fare.currencyCode)}'),
      PricingExplanationLine('${distanceKm.toStringAsFixed(1)} km'),
      PricingExplanationLine('${durationMin.round()} phút'),
      PricingExplanationLine(
        _isPeakHour(requestTime) ? 'Trong khung giờ cao điểm (BRB §2.2.12)' : 'Không trong giờ cao điểm',
      ),
      if (surge != null)
        PricingExplanationLine('Áp dụng ${surge.label}', isPositive: false)
      else
        const PricingExplanationLine('Không áp dụng Surge'),
      // No promotion engine is wired to any RPC yet — the backend never
      // returns a discount amount, so applying a voucher can only be stated
      // as a fact (the code was sent), never with a fabricated amount.
      if (voucher != null)
        PricingExplanationLine('Áp dụng mã ${voucher.code}')
      else
        const PricingExplanationLine('Không áp dụng voucher'),
    ];
    return lines;
  }

  /// Business Rule Bible v1.0 §2.2.12 Peak Hour Surcharge: 07:00–09:00 and
  /// 17:00–20:00, Monday–Friday. `DateTime.weekday` is 1=Monday..7=Sunday.
  static bool _isPeakHour(DateTime t) {
    if (t.weekday > DateTime.friday) return false;
    final minutesOfDay = t.hour * 60 + t.minute;
    const morningStart = 7 * 60, morningEnd = 9 * 60;
    const eveningStart = 17 * 60, eveningEnd = 20 * 60;
    return (minutesOfDay >= morningStart && minutesOfDay < morningEnd) ||
        (minutesOfDay >= eveningStart && minutesOfDay < eveningEnd);
  }
}
