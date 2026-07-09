import 'polyline_decoder.dart';
import 'route_leg.dart';
import 'route_point.dart';

class RouteModel {
  RouteModel({
    required this.encodedPolyline,
    required this.distanceMeters,
    required this.durationSeconds,
    this.distanceText = '',
    this.durationText = '',
    this.bounds,
    this.legs = const [],
    this.steps = const [],
  }) : decodedPolyline = decodePolyline(encodedPolyline);

  final String encodedPolyline;

  /// Decoded once in the constructor — never re-decoded per frame.
  final List<RoutePoint> decodedPolyline;

  final int distanceMeters;
  final int durationSeconds;
  final String distanceText;
  final String durationText;
  final RouteBounds? bounds;
  final List<RouteLeg> legs;
  final List<RouteStep> steps;
}
