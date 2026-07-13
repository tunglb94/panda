import 'package:flutter/foundation.dart';
import '../storage/token_storage.dart';

class AuthState extends ChangeNotifier {
  bool _isLoggedIn = false;
  String? _accessToken;
  String? _refreshToken;
  String? _driverId;

  // Captured during initialize() so ApiClient can force a logout after a
  // rejected refresh token without every call site having to thread a
  // TokenStorage instance through just for that one rare path.
  TokenStorage? _storage;

  bool get isLoggedIn => _isLoggedIn;
  String? get accessToken => _accessToken;
  String? get refreshToken => _refreshToken;
  String? get driverId => _driverId;

  // Called once during app startup before runApp — no notify needed.
  Future<void> initialize(TokenStorage storage) async {
    _storage = storage;
    _accessToken = await storage.loadToken();
    _refreshToken = await storage.loadRefreshToken();
    _driverId = await storage.loadDriverId();
    _isLoggedIn = _accessToken != null && _driverId != null;
  }

  Future<void> login({
    required String accessToken,
    required String? refreshToken,
    required String driverId,
    required TokenStorage storage,
  }) async {
    _storage = storage;
    await storage.saveToken(accessToken);
    if (refreshToken != null) await storage.saveRefreshToken(refreshToken);
    await storage.saveDriverId(driverId);
    _accessToken = accessToken;
    _refreshToken = refreshToken;
    _driverId = driverId;
    _isLoggedIn = true;
    notifyListeners();
  }

  /// Called by [ApiClient] after a successful `/api/v1/auth/refresh` —
  /// updates the in-memory + persisted tokens without touching driverId or
  /// firing a full re-login flow. No [notifyListeners] — isLoggedIn doesn't
  /// change, and nothing in the UI reads accessToken/refreshToken directly
  /// (ApiClient reads them fresh on every request).
  Future<void> updateTokens({
    required String accessToken,
    required String? refreshToken,
  }) async {
    final storage = _storage;
    if (storage != null) {
      await storage.saveToken(accessToken);
      if (refreshToken != null) await storage.saveRefreshToken(refreshToken);
    }
    _accessToken = accessToken;
    _refreshToken = refreshToken;
  }

  Future<void> logout(TokenStorage storage) async {
    _storage = storage;
    await storage.clear();
    _accessToken = null;
    _refreshToken = null;
    _driverId = null;
    _isLoggedIn = false;
    notifyListeners();
  }

  /// Called by [ApiClient] when a refresh token is rejected outright (not a
  /// network hiccup) — the session is genuinely over, so this clears state
  /// and lets GoRouter's refreshListenable redirect to the login screen,
  /// the same way a real user-initiated [logout] does.
  Future<void> forceLogout() async {
    final storage = _storage;
    if (storage != null) await storage.clear();
    _accessToken = null;
    _refreshToken = null;
    _driverId = null;
    _isLoggedIn = false;
    notifyListeners();
  }
}
