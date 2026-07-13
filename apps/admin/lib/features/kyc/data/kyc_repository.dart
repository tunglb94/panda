import 'dart:typed_data';

import '../../../core/network/api_client.dart';
import '../domain/models/driver_verification_row.dart';
import '../domain/models/kyc_detail.dart';
import '../domain/models/kyc_document_item.dart';
import '../domain/models/kyc_summary.dart';

/// Talks to the KYC review dashboard endpoints (Phần 12) — every request
/// this repository issues goes through [ApiClient], which attaches the
/// admin's Bearer token and transparently refreshes it on expiry.
class KYCRepository {
  KYCRepository(this._apiClient);

  final ApiClient _apiClient;

  Future<KYCSummary> getSummary() async {
    final json = await _apiClient.get('/api/v1/admin/verifications/summary');
    return KYCSummary.fromJson(json);
  }

  /// [status] one of pending/under_review/approved/rejected/expired.
  /// [query] case-insensitive substring match against name/phone/CCCD.
  /// [sortAsc] true = oldest first, false (default) = newest first.
  Future<List<DriverVerificationRow>> listDriverVerifications({
    required String status,
    String query = '',
    bool sortAsc = false,
  }) async {
    final params = {
      'status': status,
      'limit': '200',
      if (query.isNotEmpty) 'q': query,
      if (sortAsc) 'sort': 'asc',
    };
    final path = '/api/v1/admin/verifications/drivers?${Uri(queryParameters: params).query}';
    final json = await _apiClient.get(path);
    final list = json['verifications'] as List<dynamic>? ?? [];
    return list.map((e) => DriverVerificationRow.fromJson(e as Map<String, dynamic>)).toList();
  }

  Future<KYCDetail> getDetail(String driverId) async {
    final json = await _apiClient.get('/api/v1/admin/verifications/drivers/$driverId/detail');
    return KYCDetail.fromJson(json);
  }

  Future<void> approveDriver(String driverId) =>
      _apiClient.post('/api/v1/admin/verifications/drivers/$driverId/approve');

  Future<void> rejectDriver(String driverId, String reason) => _apiClient
      .post('/api/v1/admin/verifications/drivers/$driverId/reject', body: {'reason': reason});

  Future<void> approveVehicle(String driverId) =>
      _apiClient.post('/api/v1/admin/verifications/vehicles/$driverId/approve');

  Future<void> rejectVehicle(String driverId, String reason) => _apiClient
      .post('/api/v1/admin/verifications/vehicles/$driverId/reject', body: {'reason': reason});

  /// Best-effort selfie lookup for the list table's thumbnail column (Phần
  /// 6). Returns null if not yet uploaded — never throws, since a missing
  /// avatar is cosmetic, not an error.
  Future<KYCDocumentItem?> getSelfieDocument(String driverId) async {
    try {
      final json = await _apiClient.get('/api/v1/admin/verifications/drivers/$driverId/documents');
      final list = (json['documents'] as List<dynamic>? ?? [])
          .map((e) => KYCDocumentItem.fromJson(e as Map<String, dynamic>))
          .toList();
      for (final d in list) {
        if (d.documentType == 'selfie' && d.uploaded) return d;
      }
      return null;
    } catch (_) {
      return null;
    }
  }

  Future<Uint8List> getDocumentBytes(String documentId) =>
      _apiClient.getBytes('/api/v1/admin/verifications/documents/$documentId');

  Future<Uint8List> getDocumentsZip(String driverId) =>
      _apiClient.getBytes('/api/v1/admin/verifications/drivers/$driverId/documents.zip');
}
