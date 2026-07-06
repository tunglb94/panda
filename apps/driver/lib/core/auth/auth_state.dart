import 'package:flutter/foundation.dart';
import '../storage/token_storage.dart';

class AuthState extends ChangeNotifier {
  bool _isLoggedIn = false;
  String? _accessToken;
  String? _driverId;

  bool get isLoggedIn => _isLoggedIn;
  String? get accessToken => _accessToken;
  String? get driverId => _driverId;

  // Called once during app startup before runApp — no notify needed.
  Future<void> initialize(TokenStorage storage) async {
    _accessToken = await storage.loadToken();
    _driverId = await storage.loadDriverId();
    _isLoggedIn = _accessToken != null && _driverId != null;
  }

  Future<void> login({
    required String accessToken,
    required String driverId,
    required TokenStorage storage,
  }) async {
    await storage.saveToken(accessToken);
    await storage.saveDriverId(driverId);
    _accessToken = accessToken;
    _driverId = driverId;
    _isLoggedIn = true;
    notifyListeners();
  }

  Future<void> logout(TokenStorage storage) async {
    await storage.clear();
    _accessToken = null;
    _driverId = null;
    _isLoggedIn = false;
    notifyListeners();
  }
}
