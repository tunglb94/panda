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
    this.riderId = '',
    this.distanceKm,
    this.durationMin,
    this.vehicleType = '',
    this.tripType = '',
    this.deliveryId = '',
    this.deliveryStatus = '',
  });

  final String tripId;
  final String pickupAddress;
  final String dropoffAddress;
  final String status;
  final int finalFare;
  final String fareCurrency;

  /// The rider's identity user ID (not a display name) — needed as
  /// `ratee_id` when the driver rates the rider after trip completion.
  final String riderId;

  /// Real distance/duration echoed back by `POST .../finish` (the gateway's
  /// `FinishTrip` handler forwards `distance_km`/`duration_min` straight
  /// through — see `booking_handler.go`). Only ever populated right after
  /// this driver's own device calls `finishTrip` in this app session; a
  /// trip re-fetched later via `fetchTrip` (`GET /rides/{id}`) has no way
  /// to recover them since they are never persisted on the Trip entity.
  final double? distanceKm;
  final double? durationMin;
  final String vehicleType;

  /// Best-effort — see `TripStatusClient`'s doc comment in
  /// `booking_handler.go`. Empty means "ride" (the default) or the
  /// enrichment lookup failed this poll.
  final String tripType;
  final String deliveryId;

  /// Raw `Delivery.Status` string (CREATED/ACCEPTED/PARCEL_PICKED_UP/
  /// IN_DELIVERY/DELIVERED/COMPLETED/CANCELLED) — see `DeliveryStatus` in
  /// the rider app for the same wire values. Empty if not a delivery trip
  /// or the enrichment was unavailable this poll.
  final String deliveryStatus;

  bool get isDelivery => tripType == 'delivery';

  bool get isActive =>
      status == 'driver_assigned' ||
      status == 'driver_arrived' ||
      status == 'in_progress';

  bool get isAwaitingPayment =>
      status == 'payment_pending' || status == 'payment_success';

  ActiveTrip copyWith({
    String? status,
    int? finalFare,
    String? fareCurrency,
    String? deliveryStatus,
  }) =>
      ActiveTrip(
        tripId: tripId,
        pickupAddress: pickupAddress,
        dropoffAddress: dropoffAddress,
        status: status ?? this.status,
        finalFare: finalFare ?? this.finalFare,
        fareCurrency: fareCurrency ?? this.fareCurrency,
        riderId: riderId,
        distanceKm: distanceKm,
        durationMin: durationMin,
        vehicleType: vehicleType,
        tripType: tripType,
        deliveryId: deliveryId,
        deliveryStatus: deliveryStatus ?? this.deliveryStatus,
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
      riderId: data['rider_id'] as String? ?? '',
      tripType: data['trip_type'] as String? ?? '',
      deliveryId: data['delivery_id'] as String? ?? '',
      deliveryStatus: data['delivery_status'] as String? ?? '',
    );
  }

  Future<void> arriveAtPickup(String tripId) async {
    await _client.post('/api/v1/rides/$tripId/arrive');
  }

  Future<void> startTrip(String tripId) async {
    await _client.post('/api/v1/rides/$tripId/start');
  }

  /// Driver confirms they've physically collected the package — Delivery
  /// V1's equivalent of [startTrip]. Calls the Trip service directly (see
  /// `DeliveryHandler` in the gateway); Booking's proto has no equivalent
  /// RPC. Known gap: `AcceptDispatchOfferUseCase` (Booking) never advances
  /// `Delivery.Status` past CREATED on accept — this call can fail with a
  /// precondition error until that's fixed, independent of anything in
  /// this app.
  Future<String> pickupParcel(String tripId) async {
    final data = await _client.post('/api/v1/rides/$tripId/pickup-parcel');
    return data['delivery_status'] as String? ?? '';
  }

  Future<String> startDelivery(String tripId) async {
    final data = await _client.post('/api/v1/rides/$tripId/start-delivery');
    return data['delivery_status'] as String? ?? '';
  }

  Future<String> completeDelivery(String tripId) async {
    final data = await _client.post('/api/v1/rides/$tripId/complete-delivery');
    return data['delivery_status'] as String? ?? '';
  }

  Future<ActiveTrip> finishTrip({
    required String tripId,
    required String pickupAddress,
    required String dropoffAddress,
    required double distanceKm,
    required double durationMin,
    String riderId = '',
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
      riderId: riderId,
      // The finish response already carries these back (gateway forwards
      // `vehicle_type`/`distance_km`/`duration_min` straight from
      // `FinishedTripResponse` — see `booking_handler.go`), previously
      // parsed nowhere in this app.
      distanceKm: (data['distance_km'] as num?)?.toDouble(),
      durationMin: (data['duration_min'] as num?)?.toDouble(),
      vehicleType: data['vehicle_type'] as String? ?? '',
    );
  }
}
