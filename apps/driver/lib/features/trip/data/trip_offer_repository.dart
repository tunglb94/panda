import '../../../core/network/api_client.dart';

class TripOffer {
  const TripOffer({
    required this.tripId,
    required this.pickupAddress,
    required this.dropoffAddress,
    required this.offerExpiresAt,
    this.tripType = '',
  });

  final String tripId;
  final String pickupAddress;
  final String dropoffAddress;
  final DateTime offerExpiresAt;

  /// Best-effort — see `GetDriverOffer`'s Trip-service enrichment in
  /// `booking_handler.go` (`GetDriverCurrentOfferResponse` has no
  /// `trip_type` field on its own). Empty means "ride" (the default) or
  /// the enrichment lookup failed.
  final String tripType;

  bool get isDelivery => tripType == 'delivery';
}

class TripOfferRepository {
  TripOfferRepository({required ApiClient apiClient}) : _client = apiClient;

  final ApiClient _client;

  Future<TripOffer?> getCurrentOffer() async {
    final data = await _client.get('/api/v1/driver/current-offer');
    if (data['has_offer'] != true) return null;
    return TripOffer(
      tripId: data['trip_id'] as String,
      pickupAddress: data['pickup_address'] as String? ?? '',
      dropoffAddress: data['dropoff_address'] as String? ?? '',
      offerExpiresAt: DateTime.parse(data['offer_expires_at'] as String),
      tripType: data['trip_type'] as String? ?? '',
    );
  }

  Future<void> acceptOffer(String tripId) async {
    await _client.post('/api/v1/rides/$tripId/accept');
  }

  Future<void> rejectOffer(String tripId) async {
    await _client.post('/api/v1/rides/$tripId/reject');
  }
}
