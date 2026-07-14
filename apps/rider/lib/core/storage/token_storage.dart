import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Persists auth tokens in the platform Keychain/Keystore (via
/// flutter_secure_storage) rather than plain SharedPreferences — JWTs are
/// bearer credentials and should not sit in an unencrypted prefs file on
/// disk (Security requirement: "JWT + Refresh Token", "Không lưu OTP
/// plaintext" applies the same principle to token storage generally).
/// Public API is unchanged from the previous SharedPreferences-backed
/// implementation — every call site is unaffected by this swap.
class TokenStorage {
  static const _storage = FlutterSecureStorage();

  static const _keyToken = 'access_token';
  static const _keyRefreshToken = 'refresh_token';
  static const _keyRiderId = 'rider_id';

  Future<void> saveToken(String token) => _storage.write(key: _keyToken, value: token);

  Future<String?> loadToken() => _storage.read(key: _keyToken);

  /// The access token expires after 15 minutes (see backend
  /// identity/infrastructure/jwt/config.go) — this refresh token (30-day
  /// lifetime) is what ApiClient exchanges for a new one instead of forcing
  /// a full re-login every 15 minutes.
  Future<void> saveRefreshToken(String token) => _storage.write(key: _keyRefreshToken, value: token);

  Future<String?> loadRefreshToken() => _storage.read(key: _keyRefreshToken);

  Future<void> saveRiderId(String id) => _storage.write(key: _keyRiderId, value: id);

  Future<String?> loadRiderId() => _storage.read(key: _keyRiderId);

  Future<void> clear() => _storage.deleteAll();
}
