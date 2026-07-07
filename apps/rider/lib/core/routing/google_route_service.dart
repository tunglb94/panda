import 'dart:convert';

import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:http/http.dart' as http;

import 'polyline_decoder.dart';
import 'route_model.dart';
import 'route_service.dart';

class GoogleRouteService implements RouteService {
  const GoogleRouteService({required String apiKey}) : _apiKey = apiKey;

  final String _apiKey;

  static const _baseUrl =
      'https://maps.googleapis.com/maps/api/directions/json';

  @override
  Future<RouteModel> getRoute(LatLng origin, LatLng destination) async {
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

    final route = (data['routes'] as List).first as Map<String, dynamic>;
    final leg = (route['legs'] as List).first as Map<String, dynamic>;

    final distanceMeters =
        (leg['distance']['value'] as num).toInt();
    final durationSeconds =
        (leg['duration']['value'] as num).toInt();
    final distanceText = leg['distance']['text'] as String;
    final durationText = leg['duration']['text'] as String;

    final encodedPolyline =
        route['overview_polyline']['points'] as String;

    return RouteModel(
      polylinePoints: decodePolyline(encodedPolyline),
      distanceMeters: distanceMeters,
      durationSeconds: durationSeconds,
      distanceText: distanceText,
      durationText: durationText,
    );
  }
}
