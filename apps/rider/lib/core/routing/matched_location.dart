import 'route_point.dart';

/// Immutable snapshot of a single GPS fix matched against the active route.
///
/// When [isMatched] is true, [matchedPoint] is the perpendicular projection of
/// [originalPoint] onto the nearest route segment.  When false (GPS too far
/// from any segment), [matchedPoint] equals [originalPoint] and callers should
/// treat the position as off-route.
class MatchedLocation {
  const MatchedLocation({
    required this.originalPoint,
    required this.matchedPoint,
    required this.distanceFromRouteMeters,
    required this.nearestSegmentIndex,
    required this.progressMeters,
    required this.progressPercent,
    required this.isMatched,
    required this.timestamp,
  });

  /// Raw GPS fix before projection.
  final RoutePoint originalPoint;

  /// Projected point on the route polyline.
  ///
  /// Equals [originalPoint] when [isMatched] is false.
  final RoutePoint matchedPoint;

  /// Perpendicular distance from [originalPoint] to the nearest segment, in
  /// metres.  Always computed regardless of [isMatched].
  final double distanceFromRouteMeters;

  /// Index `i` of the polyline segment `[i] → [i+1]` nearest to the GPS fix.
  final int nearestSegmentIndex;

  /// Route-polyline distance from the route start to [matchedPoint], in metres.
  /// Computed even when [isMatched] is false so progress tracking remains
  /// consistent.
  final double progressMeters;

  /// [progressMeters] / totalRouteDistance, clamped to [0.0, 1.0].
  final double progressPercent;

  /// True when [distanceFromRouteMeters] ≤ the match threshold (default 30 m).
  final bool isMatched;

  /// Timestamp of the original GPS fix.
  final DateTime timestamp;
}
