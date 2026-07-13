import 'package:shared_preferences/shared_preferences.dart';

class TokenStorage {
  static const _keyToken = 'access_token';
  static const _keyRefreshToken = 'refresh_token';
  static const _keyAdminId = 'admin_id';

  Future<void> saveToken(String token) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyToken, token);
  }

  Future<String?> loadToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keyToken);
  }

  /// The access token expires after 15 minutes (see backend
  /// identity/infrastructure/jwt/config.go) — this refresh token (7-day
  /// lifetime) is what ApiClient exchanges for a new one instead of forcing
  /// a full re-login every 15 minutes.
  Future<void> saveRefreshToken(String token) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyRefreshToken, token);
  }

  Future<String?> loadRefreshToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keyRefreshToken);
  }

  Future<void> saveAdminId(String id) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyAdminId, id);
  }

  Future<String?> loadAdminId() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keyAdminId);
  }

  Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_keyToken);
    await prefs.remove(_keyRefreshToken);
    await prefs.remove(_keyAdminId);
  }
}
