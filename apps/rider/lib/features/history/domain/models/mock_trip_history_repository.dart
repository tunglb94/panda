import 'mock_trip_history_catalog.dart';
import 'trip_history_entry.dart';

/// Lets a caller preview all required Trip History UI states without a real
/// backend: [normal] returns the sample list, [empty] returns no trips,
/// [error] throws. Selected from the dev "Preview state" menu on
/// `TripHistoryPage`, same convention as the Notification Center (R-03).
enum TripHistoryDemoMode { normal, empty, error }

/// Mock repository for Trip History. No HTTP requests, no backend
/// dependency — see `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.1. The Trip
/// service already has a real 7-status state machine
/// (`backend/services/trip`), but nothing here calls it.
class MockTripHistoryRepository {
  const MockTripHistoryRepository();

  Future<List<TripHistoryEntry>> fetchHistory({
    TripHistoryDemoMode mode = TripHistoryDemoMode.normal,
  }) async {
    await Future.delayed(const Duration(milliseconds: 800));
    switch (mode) {
      case TripHistoryDemoMode.error:
        throw StateError('Mock error: could not load trip history (simulated).');
      case TripHistoryDemoMode.empty:
        return const [];
      case TripHistoryDemoMode.normal:
        return MockTripHistoryCatalog.sample();
    }
  }
}
