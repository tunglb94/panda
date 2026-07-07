/// Traffic conditions along the pickup leg. An enum (not a `String`) per
/// this project's state-machine convention — see `TripOfferState`.
enum TrafficLevel { normal, slow, heavy }

extension TrafficLevelX on TrafficLevel {
  String get label => switch (this) {
        TrafficLevel.normal => 'Normal traffic',
        TrafficLevel.slow => 'Slow traffic',
        TrafficLevel.heavy => 'Heavy traffic',
      };
}

/// Mock "driving to pickup" progress snapshot — no map, no GPS. [progress]
/// is the percentage of the pickup leg *remaining* (100 when navigation
/// just started, ticking down as the driver approaches; 0 means arrived —
/// reaching it is what drives `TripOfferState.navigatingToPickup ->
/// arrivedAtPickup`, Phase D-06. Phase D-05 floored at 20 before `arrived`
/// existed).
class RouteProgressModel {
  const RouteProgressModel({
    required this.remainingDistanceKm,
    required this.remainingDurationMin,
    required this.progress,
    required this.trafficLevel,
  });

  final double remainingDistanceKm;
  final double remainingDurationMin;
  final int progress;
  final TrafficLevel trafficLevel;

  /// Base pickup-leg distance/duration mirrors `TripOffer.distanceToPickupKm`
  /// (1.8 km) at an assumed ~18 km/h approach speed (6 min at 100%), so the
  /// Assigned screen's static ETA and this screen's live one agree at the
  /// start of navigation. Traffic only slows the ETA down — the physical
  /// distance remaining is unaffected by traffic.
  static RouteProgressModel mock({
    required int progress,
    required TrafficLevel trafficLevel,
  }) {
    const baseDistanceKm = 1.8;
    const baseDurationMin = 6.0;
    final trafficMultiplier = switch (trafficLevel) {
      TrafficLevel.normal => 1.0,
      TrafficLevel.slow => 1.3,
      TrafficLevel.heavy => 1.6,
    };
    final fraction = progress / 100;
    return RouteProgressModel(
      remainingDistanceKm: baseDistanceKm * fraction,
      remainingDurationMin: baseDurationMin * fraction * trafficMultiplier,
      progress: progress,
      trafficLevel: trafficLevel,
    );
  }

  /// Next mock tick as the driver gets closer — decrements [progress] by
  /// [by], floored at [floor] (0 — Arrived — since Phase D-06).
  RouteProgressModel stepDown({int by = 20, int floor = 0}) {
    final next = (progress - by).clamp(floor, 100);
    return RouteProgressModel.mock(progress: next, trafficLevel: trafficLevel);
  }
}
