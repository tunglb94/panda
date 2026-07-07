import 'package:google_maps_flutter/google_maps_flutter.dart';

class RouteModel {
  const RouteModel({
    required this.polylinePoints,
    required this.distanceMeters,
    required this.durationSeconds,
    required this.distanceText,
    required this.durationText,
  });

  final List<LatLng> polylinePoints;
  final int distanceMeters;
  final int durationSeconds;
  final String distanceText;
  final String durationText;
}
