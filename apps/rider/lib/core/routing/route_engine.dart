import 'route_model.dart';
import 'route_point.dart';
import 'route_provider.dart';

/// Single access point for all routing in the Rider App.
///
/// Callers use [loadRoute] to fetch and cache a route, [refresh] to re-fetch
/// the same endpoints, [clear] to drop the cache (e.g., when the user edits
/// pickup or destination), and [dispose] when the screen is torn down.
class RouteEngine {
  RouteEngine({required RouteProvider provider}) : _provider = provider;

  final RouteProvider _provider;

  RouteModel? _currentRoute;
  RoutePoint? _lastOrigin;
  RoutePoint? _lastDestination;
  bool _disposed = false;

  /// The most recently loaded route, or null if [clear] was called / no route
  /// has been loaded yet.
  RouteModel? get currentRoute => _currentRoute;

  /// Fetch a route from [origin] to [destination], cache it, and return it.
  Future<RouteModel> loadRoute(RoutePoint origin, RoutePoint destination) async {
    assert(!_disposed, 'RouteEngine used after dispose()');
    _lastOrigin = origin;
    _lastDestination = destination;
    _currentRoute = await _provider.calculateRoute(origin, destination);
    return _currentRoute!;
  }

  /// Re-fetch the route using the same endpoints as the last [loadRoute] call.
  /// Returns null if no route has been loaded (or after [clear]).
  Future<RouteModel?> refresh() async {
    if (_lastOrigin == null || _lastDestination == null) return null;
    return loadRoute(_lastOrigin!, _lastDestination!);
  }

  /// Drop the cached route and forget endpoints.
  ///
  /// Call when the user edits pickup, edits destination, cancels a trip,
  /// or completes a trip.
  void clear() {
    _currentRoute = null;
    _lastOrigin = null;
    _lastDestination = null;
  }

  void dispose() {
    clear();
    _disposed = true;
  }
}
