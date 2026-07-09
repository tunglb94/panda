import 'route_model.dart';
import 'route_point.dart';

/// Provider-independent interface for route calculation.
///
/// Implementations must not import google_maps_flutter or any other
/// map SDK — all coordinates flow through [RoutePoint].
abstract interface class RouteProvider {
  Future<RouteModel> calculateRoute(RoutePoint origin, RoutePoint destination);
}
