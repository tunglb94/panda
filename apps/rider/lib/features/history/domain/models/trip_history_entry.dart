import 'package:rider/features/booking/domain/models/mock_fare_calculator.dart';
import 'package:rider/features/booking/domain/models/payment_method.dart';
import 'package:rider/features/booking/domain/models/vehicle_option.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/features/trip/domain/models/mock_driver.dart';

import 'trip_history_status.dart';
import 'trip_timeline_event.dart';

/// A single past trip shown in Trip History / Trip Detail / Receipt.
///
/// Deliberately composed from types already introduced by earlier phases
/// instead of new one-off fields, per this phase's "reuse existing
/// components" requirement:
/// - [route] reuses `TripSelection` (Phase 17 / R-01) — powers `PickupCard`
///   / `DestinationCard` directly on the Detail screen.
/// - [driver] reuses `MockDriver` (R-02) — powers `DriverInfoCard` directly.
/// - [fare] reuses `MockFareBreakdown` (R-01) — powers `FareSummaryCard`
///   directly, including any promo discount line.
/// - [paymentMethod] reuses `PaymentMethod` (R-01).
/// - [vehicleCategory] reuses `VehicleCategory` (R-01), for icon parity
///   with the Vehicle Selector.
///
/// All mock data — the Trip service already has a real 7-status state
/// machine (`backend/services/trip`), but nothing here calls it (see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.1).
class TripHistoryEntry {
  const TripHistoryEntry({
    required this.id,
    required this.dateTime,
    required this.route,
    required this.status,
    required this.driver,
    required this.vehicleCategory,
    required this.fare,
    required this.paymentMethod,
    required this.distanceKm,
    required this.durationMin,
    required this.timeline,
    this.ratingGiven,
    this.cancellationReason,
  });

  final String id;
  final DateTime dateTime;
  final TripSelection route;
  final TripHistoryStatus status;
  final MockDriver driver;
  final VehicleCategory vehicleCategory;
  final MockFareBreakdown fare;
  final PaymentMethod paymentMethod;
  final double distanceKm;
  final double durationMin;
  final List<TripTimelineEvent> timeline;

  /// Stars the rider gave the driver after this trip. Null when the trip
  /// was cancelled (nothing to rate) or simply never rated.
  final double? ratingGiven;

  /// Only meaningful when [status] is `cancelled`.
  final String? cancellationReason;
}

extension TripHistoryEntryX on TripHistoryEntry {
  /// Case-insensitive match against pickup/destination address and driver
  /// name — backs the Trip History search box. Purely client-side, no
  /// backend search endpoint exists.
  bool matchesQuery(String query) {
    if (query.trim().isEmpty) return true;
    final q = query.trim().toLowerCase();
    return (route.pickupAddress?.toLowerCase().contains(q) ?? false) ||
        (route.destinationAddress?.toLowerCase().contains(q) ?? false) ||
        driver.name.toLowerCase().contains(q);
  }
}
