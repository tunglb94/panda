import '../../../core/network/api_client.dart';
import '../domain/models/payout_request.dart';

/// Driver Payout admin review — talks to the gateway's admin payout
/// endpoints (`backend/services/gateway/http/handlers/admin_wallet_handler.go`).
/// Status filtering is server-side (`ListByFilter`); driver/date filtering
/// is client-side on the already-loaded page (see PayoutsPage) — the
/// repository stays a thin 1:1 wrapper over existing endpoints, same as
/// PromotionRepository.
class PayoutRepository {
  PayoutRepository(this._apiClient);

  final ApiClient _apiClient;

  Future<List<PayoutRequest>> listPayouts({required String status}) async {
    final json = await _apiClient.get('/api/v1/admin/payouts?status=$status');
    final list = json['payout_requests'] as List<dynamic>? ?? [];
    return list.map((e) => PayoutRequest.fromJson(e as Map<String, dynamic>)).toList();
  }

  /// All requests (any status) by one driver — reuses the same endpoint
  /// with `driver_id` set and `status` omitted (ListByFilter treats an
  /// empty status as "any").
  Future<List<PayoutRequest>> listByDriver(String driverId) async {
    final json = await _apiClient.get('/api/v1/admin/payouts?driver_id=$driverId');
    final list = json['payout_requests'] as List<dynamic>? ?? [];
    return list.map((e) => PayoutRequest.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<void> approve(String id) => _apiClient.post('/api/v1/admin/payouts/$id/approve');

  Future<void> reject(String id, String reason) =>
      _apiClient.post('/api/v1/admin/payouts/$id/reject', body: {'reason': reason});

  Future<void> markPaid(String id) => _apiClient.post('/api/v1/admin/payouts/$id/paid');
}
