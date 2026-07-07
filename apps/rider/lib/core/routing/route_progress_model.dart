import 'package:google_maps_flutter/google_maps_flutter.dart';

class RouteProgressModel {
  const RouteProgressModel({
    required this.progressPercent,
    required this.remainingDistance,
    required this.remainingDuration,
    required this.isOnRoute,
    required this.nearestRoutePoint,
  });

  /// 0.0 (at start) → 1.0 (at destination).
  final double progressPercent;

  /// Remaining distance in metres.
  final int remainingDistance;

  /// Remaining duration in seconds (proportional to original route duration).
  final int remainingDuration;

  /// True when the current position is within the on-route threshold.
  final bool isOnRoute;

  /// Closest point on the route polyline to the current position.
  final LatLng nearestRoutePoint;
}
