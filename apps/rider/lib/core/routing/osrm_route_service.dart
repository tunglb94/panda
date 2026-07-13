import 'dart:convert';

import 'package:http/http.dart' as http;

import 'route_leg.dart';
import 'route_model.dart';
import 'route_point.dart';
import 'route_provider.dart';

/// [RouteProvider] backed by OSRM's public demo server
/// (https://router.project-osrm.org) — real road-network routing, no API
/// key, no billing account required.
///
/// Used instead of [GoogleRouteProvider] because the Google Cloud project's
/// Billing is currently broken/unavailable, which makes the Directions API
/// reject every request (`REQUEST_DENIED`). OSRM's polyline encoding is the
/// same format Google's Directions API uses (Encoded Polyline Algorithm
/// Format), so [RouteModel]'s existing `decodePolyline` needs no changes.
///
/// The public demo server has no uptime/rate-limit guarantee — it is
/// explicitly a community demo, not meant for heavy production traffic
/// (see https://github.com/Project-OSRM/osrm-backend/wiki/Api-usage-policy).
/// Fine for this project's current stage (the same class of free public OSM
/// service `NominatimPlacesService` already relies on for place search); a
/// self-hosted OSRM instance is the natural next step if traffic grows.
class OsrmRouteProvider implements RouteProvider {
  const OsrmRouteProvider();

  static const _baseUrl = 'https://router.project-osrm.org/route/v1/driving';

  @override
  Future<RouteModel> calculateRoute(
      RoutePoint origin, RoutePoint destination) async {
    // OSRM takes coordinates as "lon,lat" (reverse of the lat,lon order
    // used everywhere else in this codebase) — easy to get backwards.
    final coords =
        '${origin.longitude},${origin.latitude};${destination.longitude},${destination.latitude}';
    final uri = Uri.parse('$_baseUrl/$coords').replace(queryParameters: {
      'overview': 'full',
      'geometries': 'polyline',
    });

    final response = await http.get(uri);
    if (response.statusCode != 200) {
      throw Exception('OSRM error: ${response.statusCode}');
    }

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    if (data['code'] != 'Ok') {
      throw Exception('OSRM status: ${data['code']}');
    }

    // Read from the route level, not routes[0].legs[0] — with a single
    // origin/destination (no intermediate waypoints) they're numerically
    // identical, but the route level is the correct total if a multi-leg
    // waypoint route is ever added later.
    final routeData = (data['routes'] as List).first as Map<String, dynamic>;
    final distanceMeters = (routeData['distance'] as num).round();
    final durationSeconds = (routeData['duration'] as num).round();
    final encodedPolyline = routeData['geometry'] as String;

    final leg = RouteLeg(
      distanceMeters: distanceMeters,
      durationSeconds: durationSeconds,
      startPoint: origin,
      endPoint: destination,
      // OSRM step instructions are maneuver codes, not the HTML strings
      // RouteStep.instruction expects, and nothing in this app renders
      // turn-by-turn steps today — left empty rather than mistranslated.
      steps: const [],
    );

    return RouteModel(
      encodedPolyline: encodedPolyline,
      distanceMeters: distanceMeters,
      durationSeconds: durationSeconds,
      distanceText: _formatDistance(distanceMeters),
      durationText: _formatDuration(durationSeconds),
      legs: [leg],
    );
  }

  static String _formatDistance(int meters) {
    if (meters < 1000) return '$meters m';
    return '${(meters / 1000).toStringAsFixed(1)} km';
  }

  static String _formatDuration(int seconds) {
    final minutes = (seconds / 60).round();
    if (minutes < 60) return '$minutes phút';
    final hours = minutes ~/ 60;
    final rem = minutes % 60;
    return rem == 0 ? '$hours giờ' : '$hours giờ $rem phút';
  }
}
