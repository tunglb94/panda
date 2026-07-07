import 'package:flutter/foundation.dart';
import '../storage/token_storage.dart';

class AuthState extends ChangeNotifier {
  bool _isLoggedIn = false;
  String? _accessToken;
  String? _riderId;

  bool get isLoggedIn => _isLoggedIn;
  String? get accessToken => _accessToken;
  String? get riderId => _riderId;

  Future<void> initialize(TokenStorage storage) async {
    _accessToken = await storage.loadToken();
    _riderId = await storage.loadRiderId();
    _isLoggedIn = _accessToken != null && _riderId != null;
  }

  Future<void> login({
    required String accessToken,
    required String riderId,
    required TokenStorage storage,
  }) async {
    await storage.saveToken(accessToken);
    await storage.saveRiderId(riderId);
    _accessToken = accessToken;
    _riderId = riderId;
    _isLoggedIn = true;
    notifyListeners();
  }

  Future<void> logout(TokenStorage storage) async {
    await storage.clear();
    _accessToken = null;
    _riderId = null;
    _isLoggedIn = false;
    notifyListeners();
  }
}
