/// Immutable snapshot of the rider's current position along the active route.
class RouteProgress {
  const RouteProgress({
    required this.progressPercent,
    required this.travelledMeters,
    required this.remainingMeters,
    required this.completedDistanceMeters,
    required this.remainingDurationSeconds,
    required this.nearestRouteIndex,
    required this.isOnRoute,
  });

  /// 0.0 (at start) → 1.0 (at destination).
  final double progressPercent;

  /// Distance from route start to the nearest projected point, in metres.
  ///
  /// When [TripMetricsEngine] is wired (driver app integration), this can be
  /// replaced with the actual GPS path distance. Without it, equals
  /// [completedDistanceMeters].
  final int travelledMeters;

  /// Distance from the nearest projected point to the route end, in metres.
  final int remainingMeters;

  /// Route-polyline distance from start to the nearest projected point, in
  /// metres. Always derived from the polyline — never from raw GPS path.
  final int completedDistanceMeters;

  /// Proportional duration estimate: `totalDuration × (1 − progressPercent)`.
  /// Does NOT incorporate traffic or real-time ETA.
  final int remainingDurationSeconds;

  /// Index `i` of the polyline segment `[i] → [i+1]` nearest to the current
  /// position. Useful for rendering the completed portion of the route.
  final int nearestRouteIndex;

  /// True when the nearest polyline point is within 30 m of the GPS fix.
  final bool isOnRoute;
}
