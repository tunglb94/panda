import 'dart:math' as math;

import 'package:google_maps_flutter/google_maps_flutter.dart';

/// Client-side distance/duration estimate used ONLY for the mock fare
/// preview in the Booking UI.
///
/// FAIRRIDE has no route/geocoding backend yet — Phase 17 notes in
/// `.ai/memory.md` explicitly defer route calculation, ETA, and distance to
/// a future "Map Abstraction Layer" / Route Engine phase. Until that exists,
/// this straight-line (haversine) estimate is what powers the Fare Summary
/// and Vehicle Selector price previews. No network call is made.
class MockTripMetrics {
  const MockTripMetrics._();

  static const double _earthRadiusKm = 6371.0;

  /// Great-circle distance between [a] and [b], in kilometres.
  static double distanceKm(LatLng a, LatLng b) {
    final dLat = _radians(b.latitude - a.latitude);
    final dLng = _radians(b.longitude - a.longitude);
    final lat1 = _radians(a.latitude);
    final lat2 = _radians(b.latitude);

    final h = math.sin(dLat / 2) * math.sin(dLat / 2) +
        math.cos(lat1) *
            math.cos(lat2) *
            math.sin(dLng / 2) *
            math.sin(dLng / 2);
    final c = 2 * math.atan2(math.sqrt(h), math.sqrt(1 - h));
    return _earthRadiusKm * c;
  }

  /// Rough duration estimate assuming a constant average city speed.
  /// Purely illustrative until a real routing engine supplies ETAs.
  static double estimateDurationMinutes(
    double distanceKm, {
    double avgSpeedKmh = 25,
  }) {
    if (distanceKm <= 0) return 0;
    return (distanceKm / avgSpeedKmh) * 60;
  }

  static double _radians(double degrees) => degrees * math.pi / 180;
}
