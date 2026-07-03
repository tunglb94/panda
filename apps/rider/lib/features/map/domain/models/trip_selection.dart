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
  });

  final LatLng pickup;
  final LatLng destination;
  final String? pickupAddress;
  final String? destinationAddress;

  @override
  String toString() =>
      'TripSelection(pickup: ${_fmt(pickup)}, destination: ${_fmt(destination)})';

  static String _fmt(LatLng p) =>
      '${p.latitude.toStringAsFixed(5)}, ${p.longitude.toStringAsFixed(5)}';
}
