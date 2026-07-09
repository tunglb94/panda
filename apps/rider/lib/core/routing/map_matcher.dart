import 'route_model.dart';
import 'route_point.dart';
import 'route_progress_model.dart';

/// Snaps a raw GPS fix onto the nearest road in the route network.
///
/// Not implemented in Phase 32 — algorithms deferred to a future phase.
abstract interface class MapMatcher {
  Future<RoutePoint> snap(RoutePoint rawPoint, RouteModel route);
}

/// Projects a position onto a route polyline and returns the fractional
/// progress [0.0, 1.0] along the route.
///
/// Not implemented in Phase 32 — algorithms deferred to a future phase.
abstract interface class RouteProjection {
  Future<double> project(RoutePoint position, RouteModel route);
}

/// Computes full route progress (distance, duration, on-route status) from a
/// snapped position.
///
/// Not implemented in Phase 32 — algorithms deferred to a future phase.
abstract interface class RouteProgressCalculator {
  Future<RouteProgressModel> calculate({
    required RoutePoint position,
    required RouteModel route,
    required int elapsedSeconds,
  });
}
