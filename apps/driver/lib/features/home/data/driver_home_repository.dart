import '../domain/models/driver_home_summary.dart';

/// Lets a caller preview all required Home dashboard UI states without a
/// real backend: [normal] returns a fully populated summary, [empty]
/// returns one with no vehicle assigned, [error] throws. Selected from the
/// Home page's dev "Preview state" menu — same convention as
/// `apps/rider`'s Notification Center / Trip History (R-03/R-04).
enum DriverHomeDemoMode { normal, empty, error }

/// Mock repository for the Home dashboard. No HTTP requests, no backend
/// dependency — see `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.3.
class DriverHomeRepository {
  const DriverHomeRepository();

  Future<DriverHomeSummary> fetchSummary({
    DriverHomeDemoMode mode = DriverHomeDemoMode.normal,
  }) async {
    await Future.delayed(const Duration(milliseconds: 700));
    switch (mode) {
      case DriverHomeDemoMode.error:
        throw StateError('Mock error: could not load Home dashboard (simulated).');
      case DriverHomeDemoMode.empty:
        return const DriverHomeSummary(
          driverName: 'Nguyen Van A',
          rating: 4.8,
          vehicleModel: null,
          plateNumber: null,
          completedTripsToday: 0,
          earningsTodayCents: 0,
          onlineDurationMinutes: 0,
        );
      case DriverHomeDemoMode.normal:
        return const DriverHomeSummary(
          driverName: 'Nguyen Van A',
          rating: 4.8,
          vehicleModel: 'Toyota Vios',
          plateNumber: '51G-123.45',
          completedTripsToday: 7,
          earningsTodayCents: 45250,
          onlineDurationMinutes: 265,
        );
    }
  }
}
