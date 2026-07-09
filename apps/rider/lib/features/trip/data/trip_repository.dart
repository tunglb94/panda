import 'package:rider/core/network/api_client.dart';

class TripDetail {
  const TripDetail({
    required this.tripId,
    required this.status,
    required this.driverId,
    required this.finalFareCents,
    required this.currency,
  });

  final String tripId;

  /// Raw backend status string: searching, driver_assigned, driver_arrived,
  /// in_progress, completed, cancelled.
  final String status;

  final String driverId;

  /// Final fare in the smallest currency unit (e.g. cents). 0 until completed.
  final int finalFareCents;

  final String currency;
}

class TripRepository {
  const TripRepository(this._client);

  final ApiClient _client;

  Future<TripDetail> getTrip(String tripId) async {
    final body = await _client.get('/api/v1/rides/$tripId');
    return TripDetail(
      tripId: body['trip_id'] as String? ?? tripId,
      status: body['trip_status'] as String? ?? 'searching',
      driverId: body['driver_id'] as String? ?? '',
      finalFareCents: (body['final_fare'] as num?)?.toInt() ?? 0,
      currency: body['currency'] as String? ?? '',
    );
  }

  Future<void> cancelRide(String tripId) async {
    await _client.post('/api/v1/rides/$tripId/cancel');
  }

  Future<void> payRide(String tripId, {String paymentMethod = 'cash'}) async {
    await _client.post(
      '/api/v1/rides/$tripId/pay',
      body: {'payment_method': paymentMethod},
    );
  }
}
