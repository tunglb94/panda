import 'package:rider/core/network/api_client.dart';

class PromoResult {
  const PromoResult({
    required this.applied,
    required this.voucherId,
    required this.voucherCode,
    required this.voucherName,
    required this.discountAmount,
    required this.finalOrderAmount,
    required this.reason,
  });

  factory PromoResult.fromJson(Map<String, dynamic> json) => PromoResult(
        applied: json['applied'] as bool? ?? false,
        voucherId: json['voucher_id'] as String? ?? '',
        voucherCode: json['voucher_code'] as String? ?? '',
        voucherName: json['voucher_name'] as String? ?? '',
        discountAmount: (json['discount_amount'] as num?)?.toInt() ?? 0,
        finalOrderAmount: (json['final_order_amount'] as num?)?.toInt() ?? 0,
        reason: json['reason'] as String? ?? '',
      );

  final bool applied;
  final String voucherId;
  final String voucherCode;
  final String voucherName;
  final int discountAmount;
  final int finalOrderAmount;
  final String reason;
}

/// Talks to the gateway's Voucher & Promotion endpoints
/// (`POST /api/v1/promo/apply`, `GET /api/v1/rider/vouchers`). Backend is
/// the only source of truth for discount amounts — this repository never
/// computes one itself.
class PromotionRepository {
  const PromotionRepository(this._client);

  final ApiClient _client;

  Future<PromoResult> apply({
    required String code,
    required int orderAmount,
    required String serviceType,
    required String tripType,
  }) async {
    final body = await _client.post('/api/v1/promo/apply', body: {
      'code': code,
      'order_amount': orderAmount,
      'service_type': serviceType,
      'trip_type': tripType,
    });
    return PromoResult.fromJson(body);
  }

  Future<Map<String, dynamic>> myVouchers() => _client.get('/api/v1/rider/vouchers');
}
