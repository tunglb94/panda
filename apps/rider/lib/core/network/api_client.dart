import 'dart:convert';
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
  /// `package:http/testing.dart`'s `MockClient` (e.g. to simulate a
  /// timeout, an "already paid" precondition failure, or a dropped
  /// connection) without a real backend running. Production call sites
  /// never pass it, so behavior is unchanged — a fresh [http.Client] is
  /// created and reused for the lifetime of this [ApiClient].
  ApiClient({required String baseUrl, required AuthState authState, http.Client? httpClient})
      : _baseUrl = baseUrl,
        _authState = authState,
        _httpClient = httpClient ?? http.Client();

  final String _baseUrl;
  final AuthState _authState;
  final http.Client _httpClient;
  static const _timeout = Duration(seconds: 15);

  Future<Map<String, dynamic>> post(String path,
      {Map<String, dynamic>? body, Duration? timeout}) async {
    final uri = Uri.parse('$_baseUrl$path');
    final response = await _httpClient
        .post(
          uri,
          headers: _headers(),
          body: body != null ? jsonEncode(body) : null,
        )
        .timeout(timeout ?? _timeout, onTimeout: _throwTimeout);
    return _parse(response);
  }

  /// [timeout] overrides the default 15s — used by the chat feature's
  /// long-poll `GET .../messages?poll=true`, which the server deliberately
  /// holds open for up to ~25s (see notificationapp.DefaultPollTimeout on
  /// the backend) waiting for a new message before responding empty.
  Future<Map<String, dynamic>> get(String path, {Duration? timeout}) async {
    final uri = Uri.parse('$_baseUrl$path');
    final response = await _httpClient
        .get(uri, headers: _headers())
        .timeout(timeout ?? _timeout, onTimeout: _throwTimeout);
    return _parse(response);
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
