// Regression tests for ApiClient's auto-refresh-on-401 behavior — the fix
// for the bug where every request in the app started permanently failing
// with 401 fifteen minutes after login (the access token's fixed TTL),
// with no way to recover short of a full re-login.
import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:driver/core/auth/auth_state.dart';
import 'package:driver/core/network/api_client.dart';
import 'package:driver/core/storage/token_storage.dart';

http.Response _json(Map<String, dynamic> body, {int status = 200}) => http.Response(
      jsonEncode(body),
      status,
      headers: {'content-type': 'application/json'},
    );

Future<AuthState> _loggedInAuthState({String refreshToken = 'valid-refresh'}) async {
  SharedPreferences.setMockInitialValues({});
  final authState = AuthState();
  await authState.login(
    accessToken: 'expired-access',
    refreshToken: refreshToken,
    driverId: 'd1',
    storage: TokenStorage(),
  );
  return authState;
}

void main() {
  testWidgets('GET retries once with a refreshed token after a 401, and succeeds', (tester) async {
    final authState = await _loggedInAuthState();
    var refreshCalls = 0;
    var dataCalls = 0;

    final client = ApiClient(
      baseUrl: 'http://test.local',
      authState: authState,
      httpClient: MockClient((req) async {
        if (req.url.path == '/api/v1/auth/refresh') {
          refreshCalls++;
          return _json({'access_token': 'fresh-access', 'refresh_token': 'fresh-refresh'});
        }
        dataCalls++;
        if (req.headers['Authorization'] == 'Bearer expired-access') {
          return _json({'error': 'access token has expired'}, status: 401);
        }
        expect(req.headers['Authorization'], 'Bearer fresh-access');
        return _json({'status': 'ok'});
      }),
    );

    final result = await client.get('/api/v1/driver/availability');

    expect(result['status'], 'ok');
    expect(refreshCalls, 1);
    expect(dataCalls, 2); // original 401 + the retry that succeeded
    expect(authState.accessToken, 'fresh-access');
    expect(authState.isLoggedIn, isTrue);
  });

  testWidgets('Refresh token rejected by the server forces logout', (tester) async {
    final authState = await _loggedInAuthState(refreshToken: 'dead-refresh');

    final client = ApiClient(
      baseUrl: 'http://test.local',
      authState: authState,
      httpClient: MockClient((req) async {
        if (req.url.path == '/api/v1/auth/refresh') {
          return _json({'error': 'refresh token has expired'}, status: 401);
        }
        return _json({'error': 'access token has expired'}, status: 401);
      }),
    );

    await expectLater(client.get('/api/v1/driver/availability'), throwsA(isA<ApiException>()));
    expect(authState.isLoggedIn, isFalse, reason: 'a rejected refresh token means the session is genuinely over');
    expect(authState.accessToken, isNull);
  });

  testWidgets('A network failure during refresh does not force logout', (tester) async {
    final authState = await _loggedInAuthState();

    final client = ApiClient(
      baseUrl: 'http://test.local',
      authState: authState,
      httpClient: MockClient((req) async {
        if (req.url.path == '/api/v1/auth/refresh') {
          throw Exception('network unreachable');
        }
        return _json({'error': 'access token has expired'}, status: 401);
      }),
    );

    await expectLater(client.get('/api/v1/driver/availability'), throwsA(isA<ApiException>()));
    expect(authState.isLoggedIn, isTrue, reason: 'a transient network error must not log the driver out');
  });

  testWidgets('No refresh token stored means the original 401 propagates immediately', (tester) async {
    SharedPreferences.setMockInitialValues({});
    final authState = AuthState();
    await authState.login(accessToken: 'expired-access', refreshToken: null, driverId: 'd1', storage: TokenStorage());
    var calls = 0;

    final client = ApiClient(
      baseUrl: 'http://test.local',
      authState: authState,
      httpClient: MockClient((req) async {
        calls++;
        return _json({'error': 'access token has expired'}, status: 401);
      }),
    );

    await expectLater(client.get('/api/v1/driver/availability'), throwsA(isA<ApiException>()));
    expect(calls, 1, reason: 'no refresh token to try, so it must not retry at all');
  });
}
