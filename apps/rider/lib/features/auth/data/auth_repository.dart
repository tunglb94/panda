import '../../../core/device/device_info.dart';
import '../../../core/network/api_client.dart';

class LoginResult {
  const LoginResult({
    required this.accessToken,
    required this.refreshToken,
    required this.riderId,
    required this.isNewUser,
  });

  final String accessToken;
  final String? refreshToken;
  final String riderId;
  final bool isNewUser;
}

class OtpRequestResult {
  const OtpRequestResult({required this.expiresIn, this.debugOtpCode});

  final int expiresIn;

  /// Only ever non-null when the backend's APP_ENV=development — see
  /// plan's OTP dev visibility decision. Never present in production.
  final String? debugOtpCode;
}

/// Phone OTP + Google Sign-In — the "no office visit" login/signup path.
/// A correct OTP (or a verified Google account) on a phone/email the
/// backend has never seen auto-creates the account (find-or-create — see
/// backend gateway's otp_auth_handler.go).
class AuthRepository {
  const AuthRepository(this._client);

  final ApiClient _client;

  Future<OtpRequestResult> requestOtp(String phone) async {
    final body = await _client.post(
      '/api/v1/auth/otp/request',
      body: {'phone': phone},
    );
    return OtpRequestResult(
      expiresIn: (body['expires_in'] as num).toInt(),
      debugOtpCode: body['debug_otp_code'] as String?,
    );
  }

  Future<LoginResult> verifyOtp({required String phone, required String code}) async {
    final body = await _client.post(
      '/api/v1/auth/otp/verify',
      body: {'phone': phone, 'code': code, 'user_type': 'rider', ...await _deviceFields()},
    );
    return _loginResultFromBody(body);
  }

  Future<LoginResult> loginWithGoogle(String idToken) async {
    final body = await _client.post(
      '/api/v1/auth/google',
      body: {'id_token': idToken, 'user_type': 'rider', ...await _deviceFields()},
    );
    return _loginResultFromBody(body);
  }

  /// Best-effort device metadata (Device & Security phase) — the backend
  /// upserts `user_devices`/`login_history` from these; every field is
  /// optional server-side, so a collection failure here never blocks login.
  Future<Map<String, dynamic>> _deviceFields() async => {
        'device_id': await DeviceInfo.deviceId(),
        'platform': DeviceInfo.platform(),
        'model': DeviceInfo.model(),
        'app_version': DeviceInfo.appVersion(),
        'fcm_token': DeviceInfo.fcmToken(),
      };

  LoginResult _loginResultFromBody(Map<String, dynamic> body) => LoginResult(
        accessToken: body['access_token'] as String,
        refreshToken: body['refresh_token'] as String?,
        riderId: body['user_id'] as String,
        isNewUser: body['is_new_user'] as bool? ?? false,
      );
}
