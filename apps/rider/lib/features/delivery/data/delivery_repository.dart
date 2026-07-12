import 'package:rider/core/network/api_client.dart';

class DeliveryBookResult {
  const DeliveryBookResult({required this.tripId, required this.deliveryId, required this.status});

  final String tripId;

  /// Empty if the gateway's Trip-service connection wasn't available when
  /// the order was created (see `TRIP_ADDR` in `gateway/cmd/server/main.go`)
  /// — the trip itself is still created either way, only this id is best-effort.
  final String deliveryId;
  final String status;
}

/// Creates and reads delivery orders — thin wrapper reusing the same
/// `POST /api/v1/rides` / `GET /api/v1/rides/{tripId}` endpoints as Ride
/// (a delivery is a Trip with `trip_type: "delivery"` plus a linked
/// Delivery aggregate; see the backend's Delivery V1 design). No separate
/// `/api/v1/deliveries` namespace exists on the gateway.
class DeliveryRepository {
  const DeliveryRepository(this._client);

  final ApiClient _client;

  Future<DeliveryBookResult> bookDelivery({
    required String pickupAddress,
    required double pickupLat,
    required double pickupLon,
    required String receiverAddress,
    String pickupContactName = '',
    required String receiverName,
    required String receiverPhone,
    String packageNote = '',
    int packageValueCents = 0,
  }) async {
    final body = await _client.post('/api/v1/rides', body: {
      'pickup_address': pickupAddress,
      'dropoff_address': receiverAddress,
      'pickup_lat': pickupLat,
      'pickup_lon': pickupLon,
      'trip_type': 'delivery',
      'pickup_contact_name': pickupContactName,
      'receiver_name': receiverName,
      'receiver_phone': receiverPhone,
      'package_note': packageNote,
      'package_value': packageValueCents,
    });
    return DeliveryBookResult(
      tripId: body['trip_id'] as String? ?? '',
      deliveryId: body['delivery_id'] as String? ?? '',
      status: body['status'] as String? ?? '',
    );
  }
}
