import 'package:shared_preferences/shared_preferences.dart';

class TokenStorage {
  static const _keyToken = 'access_token';
  static const _keyDriverId = 'driver_id';

  Future<void> saveToken(String token) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyToken, token);
  }

  Future<String?> loadToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keyToken);
  }

  Future<void> saveDriverId(String id) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyDriverId, id);
  }

  Future<String?> loadDriverId() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keyDriverId);
  }

  Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_keyToken);
    await prefs.remove(_keyDriverId);
  }
}
