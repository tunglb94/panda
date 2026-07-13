import '../../../core/network/api_client.dart';

class AdminLoginResult {
  const AdminLoginResult({
    required this.accessToken,
    required this.refreshToken,
    required this.adminId,
  });

  final String accessToken;
  final String? refreshToken;
  final String adminId;
}

class AuthRepository {
  AuthRepository(this._apiClient);

  final ApiClient _apiClient;

  Future<AdminLoginResult> loginAdmin(String phone) async {
    final body = await _apiClient.post('/api/v1/auth/admin/login', body: {'phone': phone});
    return AdminLoginResult(
      accessToken: body['access_token'] as String,
      refreshToken: body['refresh_token'] as String?,
      adminId: body['admin_id'] as String,
    );
  }
}
