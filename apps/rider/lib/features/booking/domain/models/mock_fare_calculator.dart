import 'dart:math' as math;

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
class MockFareBreakdown {
  const MockFareBreakdown({
    required this.baseFareCents,
    required this.distanceFareCents,
    required this.timeFareCents,
    required this.bookingFeeCents,
    required this.rideFareCents,
    required this.discountCents,
    required this.totalCents,
    this.currencyCode = 'USD',
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

  /// Formats [cents] using this breakdown's currency. USD gets a `$` prefix;
  /// any other configured code is shown as a plain ISO prefix (mock-only —
  /// no locale-aware currency formatting is wired in yet).
  String format(int cents) {
    final symbol = currencyCode == 'USD' ? '\$' : '$currencyCode ';
    return '$symbol${(cents / 100).toStringAsFixed(2)}';
  }
}
