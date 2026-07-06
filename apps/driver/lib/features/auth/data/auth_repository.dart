import '../../../core/network/api_client.dart';

class LoginResult {
  const LoginResult({required this.accessToken, required this.driverId});

  final String accessToken;
  final String driverId;
}

class AuthRepository {
  const AuthRepository(this._client);

  final ApiClient _client;

  Future<LoginResult> loginDriver(String phone) async {
    final body = await _client.post(
      '/api/v1/auth/login',
      body: {'phone': phone},
    );
    return LoginResult(
      accessToken: body['access_token'] as String,
      driverId: body['driver_id'] as String,
    );
  }
}
