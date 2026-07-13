import 'package:google_maps_flutter/google_maps_flutter.dart';

/// Immutable value object representing a confirmed pickup + destination pair.
///
/// Created when the user finishes the Phase 17 selection flow.
/// [pickupAddress] and [destinationAddress] are null until a geocoding phase
/// populates them (Phase 18+).
class TripSelection {
  const TripSelection({
    required this.pickup,
    required this.destination,
    this.pickupAddress,
    this.destinationAddress,
    this.routeDistanceMeters,
    this.routeDurationSeconds,
  });

  final LatLng pickup;
  final LatLng destination;
  final String? pickupAddress;
  final String? destinationAddress;

  /// The real road-route distance/duration MapPage already fetched from
  /// Google Directions (`RouteEngine`/`RouteProvider`), NOT a straight-line
  /// estimate. Null when no route was available yet (e.g. the fetch is
  /// still in flight, failed, or this TripSelection is a demo/sample one) —
  /// callers must fall back to a straight-line estimate in that case, never
  /// pretend a real route exists.
  final int? routeDistanceMeters;
  final int? routeDurationSeconds;

  @override
  String toString() =>
      'TripSelection(pickup: ${_fmt(pickup)}, destination: ${_fmt(destination)})';

  static String _fmt(LatLng p) =>
      '${p.latitude.toStringAsFixed(5)}, ${p.longitude.toStringAsFixed(5)}';
}
