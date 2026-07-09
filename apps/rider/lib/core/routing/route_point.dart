/// A geographic coordinate — latitude and longitude in WGS-84 degrees.
///
/// Used throughout the routing layer so no package import (e.g., google_maps_flutter)
/// leaks into the domain.
class RoutePoint {
  const RoutePoint({required this.latitude, required this.longitude});

  final double latitude;
  final double longitude;

  @override
  String toString() =>
      'RoutePoint(${latitude.toStringAsFixed(6)}, ${longitude.toStringAsFixed(6)})';

  @override
  bool operator ==(Object other) =>
      other is RoutePoint &&
      other.latitude == latitude &&
      other.longitude == longitude;

  @override
  int get hashCode => Object.hash(latitude, longitude);
}

/// Axis-aligned bounding rectangle for a route.
class RouteBounds {
  const RouteBounds({required this.northeast, required this.southwest});

  final RoutePoint northeast;
  final RoutePoint southwest;
}
