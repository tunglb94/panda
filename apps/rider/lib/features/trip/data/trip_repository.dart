import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/trip/domain/models/driver_profile.dart';

class TripDetail {
  const TripDetail({
    required this.tripId,
    required this.status,
    required this.driverId,
    required this.finalFareCents,
    required this.currency,
    this.pickupAddress = '',
    this.dropoffAddress = '',
    this.dispatchStatus = '',
    this.tripType = '',
    this.deliveryId = '',
    this.deliveryStatus = '',
  });

  final String tripId;

  /// Raw backend status string: searching, driver_assigned, driver_arrived,
  /// in_progress, completed, cancelled.
  final String status;

  final String driverId;

  /// Final fare in the smallest currency unit (e.g. cents). 0 until completed.
  final int finalFareCents;

  final String currency;

  /// `GET /api/v1/rides/{tripId}` already returns these two fields (see
  /// `booking_handler.go`'s `GetBookingDetails` JSON mapping) — parsing them
  /// here means a screen no longer has to rely solely on navigation
  /// arguments passed in from a list tile to know the route.
  final String pickupAddress;
  final String dropoffAddress;

  /// Raw dispatch status ("searching"/"assigned"/... or "unknown" if no
  /// dispatch job exists yet for this trip). Empty if the backend omitted it.
  final String dispatchStatus;

  /// "ride" or "delivery". Empty means the gateway couldn't enrich this
  /// response with Trip-service data (see `TripStatusClient` in
  /// `booking_handler.go` — best-effort, never fails the request) — treat
  /// empty the same as "ride" for UI purposes, never as "unknown/delivery".
  final String tripType;

  /// Empty for a Ride trip, or when the gateway's Trip-service enrichment
  /// was unavailable.
  final String deliveryId;

  /// Raw `Delivery.Status` string: CREATED/ACCEPTED/PARCEL_PICKED_UP/
  /// IN_DELIVERY/DELIVERED/COMPLETED/CANCELLED. Empty if not a delivery
  /// trip or the enrichment was unavailable.
  final String deliveryStatus;

  bool get isDelivery => tripType == 'delivery';
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
      pickupAddress: body['pickup_address'] as String? ?? '',
      dropoffAddress: body['dropoff_address'] as String? ?? '',
      dispatchStatus: body['dispatch_status'] as String? ?? '',
      tripType: body['trip_type'] as String? ?? '',
      deliveryId: body['delivery_id'] as String? ?? '',
      deliveryStatus: body['delivery_status'] as String? ?? '',
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

  Future<void> submitRating(
    String tripId,
    int stars, {
    required String driverId,
    String? comment,
  }) async {
    final body = <String, dynamic>{
      'stars': stars,
      'ratee_id': driverId,
      'role': 'rider',
    };
    if (comment != null && comment.isNotEmpty) body['comment'] = comment;
    await _client.post('/api/v1/rides/$tripId/rate', body: body);
  }

  Future<DriverProfile> fetchDriverProfile(String driverId) async {
    final body = await _client.get('/api/v1/drivers/$driverId/profile');
    return DriverProfile(
      vehicleBrand: body['vehicle_brand'] as String? ?? '',
      vehicleModel: body['vehicle_model'] as String? ?? '',
      vehicleColor: body['vehicle_color'] as String? ?? '',
      plateNumber: body['plate_number'] as String? ?? '—',
      verificationStatus: body['verification_status'] as String? ?? '',
    );
  }
}
