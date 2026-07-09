import 'rider_trip_status.dart';
import 'mock_driver.dart';

/// Central mock data source for the Trip lifecycle UI module.
///
/// Every value here is a placeholder pending real backend wiring — see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` Rider App Roadmap stage R5 (Trip
/// tracking screen, blocked on a Driver App location broadcast and the
/// Dispatch status endpoint).
class MockTripCatalog {
  const MockTripCatalog._();

  static const MockDriver sampleDriver = MockDriver(
    name: 'Nguyen Van A',
    vehicleModel: 'Toyota Vios',
    plateNumber: '51G-123.45',
    rating: 4.8,
  );

  /// Illustrative ETA per status. No routing/dispatch backend feeds this —
  /// compare to `MockTripMetrics` in the booking module, which at least
  /// derives a straight-line estimate; here there is no live driver position
  /// to estimate from at all, so these are fixed placeholders.
  static Duration etaFor(RiderTripStatus status) => switch (status) {
        RiderTripStatus.searchingDriver => Duration.zero,
        RiderTripStatus.driverAssigned => const Duration(minutes: 5),
        RiderTripStatus.driverArriving => const Duration(minutes: 2),
        RiderTripStatus.inProgress => const Duration(minutes: 12),
        RiderTripStatus.completed => Duration.zero,
        RiderTripStatus.cancelled => Duration.zero,
      };

  static String estimatedArrivalLabel(Duration eta) {
    if (eta == Duration.zero) return '--:--';
    final arrival = DateTime.now().add(eta);
    final hh = arrival.hour.toString().padLeft(2, '0');
    final mm = arrival.minute.toString().padLeft(2, '0');
    return '$hh:$mm';
  }
}
