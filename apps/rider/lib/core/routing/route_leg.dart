import 'route_point.dart';

class RouteLeg {
  const RouteLeg({
    required this.distanceMeters,
    required this.durationSeconds,
    required this.startPoint,
    required this.endPoint,
    this.steps = const [],
  });

  final int distanceMeters;
  final int durationSeconds;
  final RoutePoint startPoint;
  final RoutePoint endPoint;
  final List<RouteStep> steps;
}

class RouteStep {
  const RouteStep({
    required this.distanceMeters,
    required this.durationSeconds,
    required this.startPoint,
    required this.endPoint,
    required this.instruction,
  });

  final int distanceMeters;
  final int durationSeconds;
  final RoutePoint startPoint;
  final RoutePoint endPoint;
  final String instruction;
}
