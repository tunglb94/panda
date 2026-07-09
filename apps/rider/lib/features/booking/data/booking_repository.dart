import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

class BookResult {
  const BookResult({required this.tripId, required this.status});

  final String tripId;
  final String status;
}

class BookingRepository {
  const BookingRepository(this._client);

  final ApiClient _client;

  Future<BookResult> bookRide(TripSelection selection) async {
    // Geocoding is not yet implemented; fall back to coordinate strings so the
    // backend's non-empty address validation always passes.
    final pickupAddr = selection.pickupAddress?.isNotEmpty == true
        ? selection.pickupAddress!
        : _coords(selection.pickup.latitude, selection.pickup.longitude);
    final dropoffAddr = selection.destinationAddress?.isNotEmpty == true
        ? selection.destinationAddress!
        : _coords(selection.destination.latitude, selection.destination.longitude);

    final body = await _client.post(
      '/api/v1/rides',
      body: {
        'pickup_address': pickupAddr,
        'dropoff_address': dropoffAddr,
        'pickup_lat': selection.pickup.latitude,
        'pickup_lon': selection.pickup.longitude,
      },
    );
    return BookResult(
      tripId: body['trip_id'] as String,
      status: body['status'] as String,
    );
  }

  static String _coords(double lat, double lon) =>
      '${lat.toStringAsFixed(5)}, ${lon.toStringAsFixed(5)}';
}
