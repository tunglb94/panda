import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

class BookResult {
  const BookResult({
    required this.tripId,
    required this.status,
    this.voucherCode,
    this.discountAmount,
  });

  final String tripId;
  final String status;

  /// Set only when the backend actually redeemed a voucher against this
  /// trip (see gateway's BookingHandler.redeemVoucher) — null means no
  /// voucher was applied, not that the field is unknown.
  final String? voucherCode;
  final int? discountAmount;
}

class BookingRepository {
  const BookingRepository(this._client);

  final ApiClient _client;

  /// [voucherCode]/[orderAmount] are set together when the rider applied a
  /// voucher in the booking screen (`PromotionRepository.apply` already
  /// showed them the discount computed against this same [orderAmount], the
  /// estimated fare) — the backend re-validates and re-computes the
  /// discount itself rather than trusting any client-supplied amount.
  Future<BookResult> bookRide(
    TripSelection selection, {
    String? voucherCode,
    int? orderAmount,
  }) async {
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
        if (voucherCode != null && voucherCode.isNotEmpty) 'voucher_code': voucherCode,
        if (orderAmount != null && orderAmount > 0) 'order_amount': orderAmount,
      },
    );
    return BookResult(
      tripId: body['trip_id'] as String,
      status: body['status'] as String,
      voucherCode: body['voucher_code'] as String?,
      discountAmount: (body['discount_amount'] as num?)?.toInt(),
    );
  }

  static String _coords(double lat, double lon) =>
      '${lat.toStringAsFixed(5)}, ${lon.toStringAsFixed(5)}';
}
