import 'dart:convert';
import 'dart:typed_data';
import 'package:http/http.dart' as http;
import '../auth/auth_state.dart';

class ApiException implements Exception {
  const ApiException({required this.statusCode, required this.message});

  final int statusCode;
  final String message;

  @override
  String toString() => 'ApiException($statusCode): $message';
}

class ApiClient {
  /// [httpClient] is injectable so tests can substitute
  /// `package:http/testing.dart`'s `MockClient` without a real backend
  /// running. Mirrors `apps/driver`/`apps/rider`'s `ApiClient`.
  ApiClient({required String baseUrl, required AuthState authState, http.Client? httpClient})
      : _baseUrl = baseUrl,
        _authState = authState,
        _httpClient = httpClient ?? http.Client();

  final String _baseUrl;
  final AuthState _authState;
  final http.Client _httpClient;
  static const _timeout = Duration(seconds: 15);

  // A single in-flight refresh shared by every concurrent request that hits
  // 401 at the same time — without this, N simultaneous requests each
  // expiring together would fire N separate refresh calls.
  Future<bool>? _refreshInFlight;

  Future<Map<String, dynamic>> post(String path, {Map<String, dynamic>? body}) async {
    final uri = Uri.parse('$_baseUrl$path');
    final response = await _sendWithAuthRetry(
      () => _httpClient
          .post(uri, headers: _headers(), body: body != null ? jsonEncode(body) : null)
          .timeout(_timeout, onTimeout: _throwTimeout),
    );
    return _parse(response);
  }

  Future<Map<String, dynamic>> get(String path) async {
    final uri = Uri.parse('$_baseUrl$path');
    final response = await _sendWithAuthRetry(
      () => _httpClient.get(uri, headers: _headers()).timeout(_timeout, onTimeout: _throwTimeout),
    );
    return _parse(response);
  }

  Future<Map<String, dynamic>> put(String path, {Map<String, dynamic>? body}) async {
    final uri = Uri.parse('$_baseUrl$path');
    final response = await _sendWithAuthRetry(
      () => _httpClient
          .put(uri, headers: _headers(), body: body != null ? jsonEncode(body) : null)
          .timeout(_timeout, onTimeout: _throwTimeout),
    );
    return _parse(response);
  }

  Future<Map<String, dynamic>> delete(String path) async {
    final uri = Uri.parse('$_baseUrl$path');
    final response = await _sendWithAuthRetry(
      () => _httpClient.delete(uri, headers: _headers()).timeout(_timeout, onTimeout: _throwTimeout),
    );
    return _parse(response);
  }

  /// Fetches raw bytes with the same auth-retry behavior as [get] — used for
  /// KYC document images (Phần 2/13) and the ZIP download (Phần 3), neither
  /// of which is JSON.
  Future<Uint8List> getBytes(String path) async {
    final uri = Uri.parse('$_baseUrl$path');
    final response = await _sendWithAuthRetry(
      () => _httpClient.get(uri, headers: _headers()).timeout(_timeout, onTimeout: _throwTimeout),
    );
    if (response.statusCode >= 200 && response.statusCode < 300) {
      return response.bodyBytes;
    }
    throw ApiException(statusCode: response.statusCode, message: 'Yêu cầu thất bại');
  }

  /// Runs [request] once; on a 401 (access token expired — it has a fixed
  /// 15-minute lifetime, see backend identity/infrastructure/jwt/config.go)
  /// attempts exactly one token refresh and retries [request] once more with
  /// the refreshed token.
  Future<http.Response> _sendWithAuthRetry(Future<http.Response> Function() request) async {
    final response = await request();
    if (response.statusCode != 401) return response;
    if (_authState.refreshToken == null) return response;

    final refreshed = await _refreshAccessToken();
    if (!refreshed) return response;
    return request();
  }

  /// Shares one in-flight refresh across concurrent callers and forces a
  /// logout only when the server explicitly rejects the refresh token
  /// (session genuinely over) — a network error/timeout here just fails
  /// this one request, since the refresh token itself may still be valid.
  Future<bool> _refreshAccessToken() {
    return _refreshInFlight ??= () async {
      try {
        final refreshToken = _authState.refreshToken;
        if (refreshToken == null) return false;
        final uri = Uri.parse('$_baseUrl/api/v1/auth/refresh');
        final response = await _httpClient
            .post(
              uri,
              headers: const {'Content-Type': 'application/json'},
              body: jsonEncode({'refresh_token': refreshToken}),
            )
            .timeout(_timeout, onTimeout: _throwTimeout);

        if (response.statusCode == 401) {
          await _authState.forceLogout();
          return false;
        }
        if (response.statusCode != 200) return false;

        final data = jsonDecode(response.body) as Map<String, dynamic>;
        final newAccessToken = data['access_token'] as String?;
        if (newAccessToken == null) return false;
        await _authState.updateTokens(
          accessToken: newAccessToken,
          refreshToken: data['refresh_token'] as String?,
        );
        return true;
      } catch (_) {
        return false;
      }
    }().whenComplete(() => _refreshInFlight = null);
  }

  Never _throwTimeout() => throw const ApiException(
        statusCode: 0,
        message: 'Hết thời gian chờ. Kiểm tra kết nối và thử lại.',
      );

  Map<String, String> _headers() {
    final token = _authState.accessToken;
    return {
      'Content-Type': 'application/json',
      if (token != null) 'Authorization': 'Bearer $token',
    };
  }

  Map<String, dynamic> _parse(http.Response response) {
    if (response.statusCode >= 200 && response.statusCode < 300) {
      return jsonDecode(response.body) as Map<String, dynamic>;
    }
    String message = 'Yêu cầu thất bại';
    try {
      final body = jsonDecode(response.body) as Map<String, dynamic>;
      message = body['error'] as String? ?? message;
    } catch (_) {}
    throw ApiException(statusCode: response.statusCode, message: message);
  }
}
