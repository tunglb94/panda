import 'rider_trip_status.dart';

/// Mock repository driving the rider-facing trip lifecycle animation.
///
/// This is a stand-in for what will eventually be a real subscription to the
/// Dispatch/Trip status endpoints (see Rider App Roadmap stage R5). It makes
/// no HTTP calls and has no backend dependency — it simply emits each
/// [RiderTripStatus] in order, pausing between emissions so the UI has time
/// to animate the transition.
class MockTripRepository {
  const MockTripRepository({
    this.stageDurations = const [
      Duration(seconds: 4), // time spent searching
      Duration(seconds: 3), // time spent assigned, before arriving
      Duration(seconds: 4), // time spent arriving, before trip starts
      Duration(seconds: 6), // time spent in progress, before completion
    ],
  });

  /// One entry per transition *out of* a non-terminal status, in
  /// [RiderTripStatus] declaration order. The final status (`completed`) has
  /// no further transition, so this list is one shorter than the enum.
  final List<Duration> stageDurations;

  /// Emits every [RiderTripStatus] in sequence, pausing [stageDurations]
  /// between each. Completes after emitting [RiderTripStatus.completed].
  Stream<RiderTripStatus> watchLifecycle() async* {
    final statuses = RiderTripStatus.values;
    for (var i = 0; i < statuses.length; i++) {
      yield statuses[i];
      if (i < stageDurations.length) {
        await Future.delayed(stageDurations[i]);
      }
    }
  }
}
