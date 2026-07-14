import '../../../core/network/api_client.dart';
import '../domain/models/rider_verification.dart';

/// Rider KYC — deliberately minimal (5 fields: họ tên, ngày sinh, CCCD, 2
/// ảnh) compared to Driver KYC. Mirrors `apps/driver`'s `KYCRepository`
/// conventions: `getVerification` returns null (not an error) on 404 —
/// "nothing submitted yet" is a normal state, not a failure.
class KYCRepository {
  const KYCRepository(this._client);

  final ApiClient _client;

  Future<RiderVerification?> getVerification() async {
    try {
      final body = await _client.get('/api/v1/rider/verification');
      return RiderVerification.fromJson(body);
    } on ApiException catch (e) {
      if (e.statusCode == 404) return null;
      rethrow;
    }
  }

  Future<void> uploadDocument({
    required String documentType,
    required List<int> bytes,
    required String filename,
  }) async {
    await _client.postMultipart(
      '/api/v1/rider/verification/documents',
      fileFieldName: 'file',
      bytes: bytes,
      filename: filename,
      fields: {'document_type': documentType},
    );
  }

  Future<RiderVerification> submit({
    required String fullName,
    required DateTime dateOfBirth,
    required String nationalIdNumber,
  }) async {
    final body = await _client.post('/api/v1/rider/verification', body: {
      'full_name': fullName,
      'date_of_birth': _dateOnly(dateOfBirth),
      'national_id_number': nationalIdNumber,
    });
    return RiderVerification.fromJson(body);
  }

  static String _dateOnly(DateTime d) =>
      '${d.year.toString().padLeft(4, '0')}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';
}
