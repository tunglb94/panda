import 'package:shared_preferences/shared_preferences.dart';

import '../../../core/network/api_client.dart';

class ActiveTrip {
  const ActiveTrip({
    required this.tripId,
    required this.pickupAddress,
    required this.dropoffAddress,
    required this.status,
    this.finalFare = 0,
    this.fareCurrency = '',
  });

  final String tripId;
  final String pickupAddress;
  final String dropoffAddress;
  final String status;
  final int finalFare;
  final String fareCurrency;

  bool get isActive =>
      status == 'driver_assigned' || status == 'in_progress';

  bool get isAwaitingPayment =>
      status == 'payment_pending' || status == 'payment_success';

  ActiveTrip copyWith({
    String? status,
    int? finalFare,
    String? fareCurrency,
  }) =>
      ActiveTrip(
        tripId: tripId,
        pickupAddress: pickupAddress,
        dropoffAddress: dropoffAddress,
        status: status ?? this.status,
        finalFare: finalFare ?? this.finalFare,
        fareCurrency: fareCurrency ?? this.fareCurrency,
      );
}

class ActiveTripRepository {
  ActiveTripRepository({required ApiClient apiClient}) : _client = apiClient;

  final ApiClient _client;
  static const _keyActiveTripId = 'active_trip_id';

  Future<String?> getStoredTripId() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keyActiveTripId);
  }

  Future<void> saveActiveTripId(String tripId) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyActiveTripId, tripId);
  }

  Future<void> clearActiveTripId() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_keyActiveTripId);
  }

  Future<ActiveTrip> fetchTrip(String tripId) async {
    final data = await _client.get('/api/v1/rides/$tripId');
    return ActiveTrip(
      tripId: data['trip_id'] as String,
      pickupAddress: data['pickup_address'] as String? ?? '',
      dropoffAddress: data['dropoff_address'] as String? ?? '',
      status: data['trip_status'] as String? ?? '',
      finalFare: (data['final_fare'] as num?)?.toInt() ?? 0,
      fareCurrency: data['currency'] as String? ?? '',
    );
  }

  Future<void> startTrip(String tripId) async {
    await _client.post('/api/v1/rides/$tripId/start');
  }

  Future<ActiveTrip> finishTrip({
    required String tripId,
    required String pickupAddress,
    required String dropoffAddress,
    required double distanceKm,
    required double durationMin,
  }) async {
    final data = await _client.post(
      '/api/v1/rides/$tripId/finish',
      body: {
        'vehicle_type': 'car',
        'distance_km': distanceKm,
        'duration_min': durationMin,
      },
    );
    return ActiveTrip(
      tripId: tripId,
      pickupAddress: pickupAddress,
      dropoffAddress: dropoffAddress,
      status: data['status'] as String? ?? 'completed',
      finalFare: (data['final_fare'] as num?)?.toInt() ?? 0,
      fareCurrency: data['currency'] as String? ?? '',
    );
  }
}
