import '../../../core/network/api_client.dart';

class LoginResult {
  const LoginResult({required this.accessToken, required this.riderId});

  final String accessToken;
  final String riderId;
}

class AuthRepository {
  const AuthRepository(this._client);

  final ApiClient _client;

  Future<LoginResult> loginRider(String phone) async {
    final body = await _client.post(
      '/api/v1/auth/rider/login',
      body: {'phone': phone},
    );
    return LoginResult(
      accessToken: body['access_token'] as String,
      riderId: body['rider_id'] as String,
    );
  }
}
