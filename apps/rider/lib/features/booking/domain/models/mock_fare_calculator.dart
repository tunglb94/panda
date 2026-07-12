import 'dart:math' as math;

import 'package:rider/shared/utils/currency_format.dart';

import 'vehicle_option.dart';

/// Client-side fare estimate powering the Booking UI's Fare Summary card.
///
/// Mirrors the calculation shape of
/// `backend/services/pricing/app/fare_calculator.go`
/// (`rideFare = max(base + distance*perKm + time*perMinute, minimum)`,
/// `total = rideFare + bookingFee`) so the mock preview lands in the same
/// ballpark as what the real Pricing service will later return. This
/// performs no network call — see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` Rider App Roadmap stage R4 for
/// when real fare estimation gets wired in.
///
/// The `*Cents` field names are legacy from when this mirrored USD test
/// rates; values are now whole VND (no subunit — see [formatMoney]), not
/// actually cents.
class MockFareBreakdown {
  const MockFareBreakdown({
    required this.baseFareCents,
    required this.distanceFareCents,
    required this.timeFareCents,
    required this.bookingFeeCents,
    required this.rideFareCents,
    required this.discountCents,
    required this.totalCents,
    this.currencyCode = 'VND',
  });

  final int baseFareCents;
  final int distanceFareCents;
  final int timeFareCents;
  final int bookingFeeCents;
  final int rideFareCents;
  final int discountCents;
  final int totalCents;
  final String currencyCode;

  static MockFareBreakdown calculate({
    required VehicleOption vehicle,
    required double distanceKm,
    required double durationMin,
    int discountPercent = 0,
  }) {
    final distanceFare = (vehicle.perKmCents * distanceKm).round();
    final timeFare = (vehicle.perMinuteCents * durationMin).round();
    final rideFareRaw = vehicle.baseFareCents + distanceFare + timeFare;
    final rideFare = math.max(rideFareRaw, vehicle.minimumFareCents);
    final subtotal = rideFare + vehicle.bookingFeeCents;
    final discount = ((subtotal * discountPercent) / 100).round();
    final total = (subtotal - discount).clamp(0, subtotal);

    return MockFareBreakdown(
      baseFareCents: vehicle.baseFareCents,
      distanceFareCents: distanceFare,
      timeFareCents: timeFare,
      bookingFeeCents: vehicle.bookingFeeCents,
      rideFareCents: rideFare,
      discountCents: discount,
      totalCents: total,
    );
  }

  /// Formats [amount] using this breakdown's currency.
  String format(int amount) => formatMoney(amount, currencyCode);
}
