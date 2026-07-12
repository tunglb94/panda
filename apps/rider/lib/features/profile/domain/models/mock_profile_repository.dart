import 'rider_profile.dart';

/// Sample data returned by [MockProfileRepository].
class MockRiderProfileCatalog {
  const MockRiderProfileCatalog._();

  static const RiderProfile sample = RiderProfile(
    fullName: 'Nguyen Van Nam',
    phoneNumber: '+1 555 000 0000',
    memberLevel: MemberLevel.gold,
    rating: 4.9,
    totalCompletedTrips: 128,
  );
}

/// Mock repository for the Profile screen.
///
/// Simulates a network round-trip with a short delay so the screen has a
/// genuine Loading → Success transition to animate, without making any real
/// HTTP request. [simulateError] lets a caller exercise the Error UI state
/// on demand (used by the Profile screen's dev "simulate error" action).
class MockProfileRepository {
  const MockProfileRepository();

  Future<RiderProfile> fetchProfile({bool simulateError = false}) async {
    await Future.delayed(const Duration(milliseconds: 700));
    if (simulateError) {
      throw StateError('Mock error: could not load profile (simulated).');
    }
    return MockRiderProfileCatalog.sample;
  }
}
