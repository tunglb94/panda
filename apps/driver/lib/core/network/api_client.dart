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
  ApiClient({required String baseUrl, required AuthState authState})
      : _baseUrl = baseUrl,
        _authState = authState;

  final String _baseUrl;
  final AuthState _authState;
  static const _timeout = Duration(seconds: 15);

  Future<Map<String, dynamic>> post(String path,
      {Map<String, dynamic>? body}) async {
    final uri = Uri.parse('$_baseUrl$path');
    final response = await http
        .post(
          uri,
          headers: _headers(),
          body: body != null ? jsonEncode(body) : null,
        )
        .timeout(_timeout, onTimeout: _throwTimeout);
    return _parse(response);
  }

  Future<Map<String, dynamic>> get(String path) async {
    final uri = Uri.parse('$_baseUrl$path');
    final response = await http
        .get(uri, headers: _headers())
        .timeout(_timeout, onTimeout: _throwTimeout);
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
