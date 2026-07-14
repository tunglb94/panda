import '../../../core/network/api_client.dart';
import '../domain/models/voucher.dart';

/// Voucher & Promotion CRUD — talks to the gateway's admin voucher endpoints
/// (`backend/services/gateway/http/handlers/admin_promotion_handler.go`).
class PromotionRepository {
  PromotionRepository(this._apiClient);

  final ApiClient _apiClient;

  Future<List<Voucher>> listVouchers() async {
    final json = await _apiClient.get('/api/v1/admin/vouchers');
    final list = json['vouchers'] as List<dynamic>? ?? [];
    return list.map((e) => Voucher.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<Voucher> getVoucher(String id) async {
    final json = await _apiClient.get('/api/v1/admin/vouchers/$id');
    return Voucher.fromJson(json);
  }

  Future<Voucher> createVoucher(Map<String, dynamic> form) async {
    final json = await _apiClient.post('/api/v1/admin/vouchers', body: form);
    return Voucher.fromJson(json);
  }

  Future<Voucher> updateVoucher(String id, Map<String, dynamic> form) async {
    final json = await _apiClient.put('/api/v1/admin/vouchers/$id', body: form);
    return Voucher.fromJson(json);
  }

  Future<void> enableVoucher(String id) => _apiClient.post('/api/v1/admin/vouchers/$id/enable');

  Future<void> disableVoucher(String id) => _apiClient.post('/api/v1/admin/vouchers/$id/disable');

  Future<void> deleteVoucher(String id) => _apiClient.delete('/api/v1/admin/vouchers/$id');
}
