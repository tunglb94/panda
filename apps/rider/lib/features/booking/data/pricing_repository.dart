import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/core/network/api_client.dart';

import '../domain/models/fare_estimate.dart';

/// Calls the backend's real fare estimate endpoint
/// (`POST /api/v1/rides/estimate-fare`, gateway's `PricingHandler`) — the
/// only source of truth for fare in the Rider app. No fallback/mock
/// computation exists anywhere in this repository; a failed call must
/// surface as an error to the caller, never a locally-computed number.
class PricingRepository {
  const PricingRepository(this._client);

  final ApiClient _client;

  Future<FareEstimate> estimateFare({
    required LatLng pickup,
    required LatLng destination,
    required String serviceType,
    required String tripType,
    String promoCode = '',
  }) async {
    final body = await _client.post(
      '/api/v1/rides/estimate-fare',
      body: {
        'pickup_lat': pickup.latitude,
        'pickup_lon': pickup.longitude,
        'destination_lat': destination.latitude,
        'destination_lon': destination.longitude,
        'service_type': serviceType,
        'trip_type': tripType,
        'promo_code': promoCode,
      },
    );
    return FareEstimate.fromJson(body);
  }
}
