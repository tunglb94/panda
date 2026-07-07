import 'package:shared_preferences/shared_preferences.dart';

class TokenStorage {
  static const _keyToken = 'access_token';
  static const _keyRiderId = 'rider_id';

  Future<void> saveToken(String token) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyToken, token);
  }

  Future<String?> loadToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keyToken);
  }

  Future<void> saveRiderId(String id) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyRiderId, id);
  }

  Future<String?> loadRiderId() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_keyRiderId);
  }

  Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_keyToken);
    await prefs.remove(_keyRiderId);
  }
}
