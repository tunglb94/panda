import 'dart:convert';

import 'package:http/http.dart' as http;

import 'route_leg.dart';
import 'route_model.dart';
import 'route_point.dart';
import 'route_provider.dart';

class GoogleRouteProvider implements RouteProvider {
  const GoogleRouteProvider({required String apiKey}) : _apiKey = apiKey;

  final String _apiKey;

  static const _baseUrl =
      'https://maps.googleapis.com/maps/api/directions/json';

  @override
  Future<RouteModel> calculateRoute(
      RoutePoint origin, RoutePoint destination) async {
    final uri = Uri.parse(_baseUrl).replace(queryParameters: {
      'origin': '${origin.latitude},${origin.longitude}',
      'destination': '${destination.latitude},${destination.longitude}',
      'key': _apiKey,
    });

    final response = await http.get(uri);
    if (response.statusCode != 200) {
      throw Exception('Directions API error: ${response.statusCode}');
    }

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    final status = data['status'] as String;
    if (status != 'OK') {
      throw Exception('Directions API status: $status');
    }

    final routeData = (data['routes'] as List).first as Map<String, dynamic>;
    final leg = (routeData['legs'] as List).first as Map<String, dynamic>;

    final distanceMeters = (leg['distance']['value'] as num).toInt();
    final durationSeconds = (leg['duration']['value'] as num).toInt();
    final distanceText = leg['distance']['text'] as String;
    final durationText = leg['duration']['text'] as String;
    final encodedPolyline =
        routeData['overview_polyline']['points'] as String;

    final boundsData = routeData['bounds'] as Map<String, dynamic>?;
    final bounds = boundsData != null
        ? RouteBounds(
            northeast: RoutePoint(
              latitude: (boundsData['northeast']['lat'] as num).toDouble(),
              longitude: (boundsData['northeast']['lng'] as num).toDouble(),
            ),
            southwest: RoutePoint(
              latitude: (boundsData['southwest']['lat'] as num).toDouble(),
              longitude: (boundsData['southwest']['lng'] as num).toDouble(),
            ),
          )
        : null;

    final legStartLat = (leg['start_location']['lat'] as num).toDouble();
    final legStartLng = (leg['start_location']['lng'] as num).toDouble();
    final legEndLat = (leg['end_location']['lat'] as num).toDouble();
    final legEndLng = (leg['end_location']['lng'] as num).toDouble();

    final steps = ((leg['steps'] as List?) ?? []).map((s) {
      final step = s as Map<String, dynamic>;
      return RouteStep(
        distanceMeters: (step['distance']['value'] as num).toInt(),
        durationSeconds: (step['duration']['value'] as num).toInt(),
        startPoint: RoutePoint(
          latitude: (step['start_location']['lat'] as num).toDouble(),
          longitude: (step['start_location']['lng'] as num).toDouble(),
        ),
        endPoint: RoutePoint(
          latitude: (step['end_location']['lat'] as num).toDouble(),
          longitude: (step['end_location']['lng'] as num).toDouble(),
        ),
        instruction: step['html_instructions'] as String? ?? '',
      );
    }).toList();

    final parsedLeg = RouteLeg(
      distanceMeters: distanceMeters,
      durationSeconds: durationSeconds,
      startPoint: RoutePoint(latitude: legStartLat, longitude: legStartLng),
      endPoint: RoutePoint(latitude: legEndLat, longitude: legEndLng),
      steps: steps,
    );

    // RouteModel constructor decodes encodedPolyline once and caches the result.
    return RouteModel(
      encodedPolyline: encodedPolyline,
      distanceMeters: distanceMeters,
      durationSeconds: durationSeconds,
      distanceText: distanceText,
      durationText: durationText,
      bounds: bounds,
      legs: [parsedLeg],
      steps: steps,
    );
  }
}
