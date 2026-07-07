import '../domain/models/dispatch_accept_result.dart';
import '../domain/models/mock_trip_offer_catalog.dart';
import '../domain/models/route_progress_model.dart';
import '../domain/models/trip_offer.dart';

/// Lets a caller preview all required Trips-tab data-loading UI states
/// without a real backend: [normal] returns the sample offer, [empty]
/// returns no offer (`null` — nothing incoming right now), [error] throws.
/// Selected from the Trips page's dev "Preview state" menu — same
/// convention as the Home dashboard (D-02) and `apps/rider`'s Notification
/// Center / Trip History (R-03/R-04).
enum DriverTripOfferDemoMode { normal, empty, error }

/// Mock repository for incoming trip offers. No HTTP requests, no backend
/// dependency — see `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.3. [delay] is
/// configurable per the task's "return mock offers with configurable
/// delay" requirement; callers needing a different delay (e.g. tests) can
/// override it.
class DriverTripOfferRepository {
  const DriverTripOfferRepository();

  Future<TripOffer?> fetchOffer({
    DriverTripOfferDemoMode mode = DriverTripOfferDemoMode.normal,
    Duration delay = const Duration(milliseconds: 700),
  }) async {
    await Future.delayed(delay);
    switch (mode) {
      case DriverTripOfferDemoMode.error:
        throw StateError('Mock error: could not load trip offer (simulated).');
      case DriverTripOfferDemoMode.empty:
        return null;
      case DriverTripOfferDemoMode.normal:
        return MockTripOfferCatalog.sample;
    }
  }

  /// Mock dispatch confirmation for an accepted offer. [outcome] lets a
  /// caller (the Trips page's dev "Accept outcome" menu, or a test) force
  /// which `DispatchAcceptResult` comes back; it does not call any API.
  Future<DispatchAcceptResult> acceptOffer({
    Duration delay = const Duration(milliseconds: 1200),
    DispatchAcceptStatus outcome = DispatchAcceptStatus.success,
  }) async {
    await Future.delayed(delay);
    return DispatchAcceptResult(status: outcome);
  }

  /// Mock "driving to pickup" snapshot for the Navigation screen (Phase
  /// D-05). [traffic] lets the dev "Traffic" menu (or a test) force Normal /
  /// Slow / Heavy without depending on a real map/GPS provider.
  Future<RouteProgressModel> fetchRouteProgress({
    Duration delay = const Duration(milliseconds: 600),
    TrafficLevel traffic = TrafficLevel.normal,
  }) async {
    await Future.delayed(delay);
    return RouteProgressModel.mock(progress: 100, trafficLevel: traffic);
  }
}
