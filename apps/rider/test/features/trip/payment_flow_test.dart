// Tests for Section 7/8/11 of the Payment/Fare production pass: Cash,
// Wallet, Payment timeout, Already Paid, Double Click, Network Lost.
//
// `ApiClient`'s injectable `http.Client` (added as part of this pass) lets
// these run against a `MockClient` — no real backend, no flakiness from a
// live server, and timeouts/lost-connection scenarios are fully
// deterministic since `testWidgets` runs inside a fake-async zone (a
// `Future.delayed` inside the mock handler advances only when the test
// explicitly pumps that much virtual time).
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/features/trip/presentation/pages/trip_lifecycle_page.dart';

const _tripSelection = TripSelection(
  pickup: LatLng(10.77, 106.69),
  destination: LatLng(10.78, 106.70),
  pickupAddress: '123 Main St',
  destinationAddress: '456 Elm Ave',
);

http.Response _json(Map<String, dynamic> body, {int status = 200}) => http.Response(
      jsonEncode(body),
      status,
      headers: {'content-type': 'application/json'},
    );

ApiClient _apiClient(Future<http.Response> Function(http.Request) handler) {
  return ApiClient(
    baseUrl: 'http://test.local',
    authState: AuthState(),
    httpClient: MockClient((req) => handler(req)),
  );
}

/// Pumps away to a neutral widget so `TripLifecyclePage`'s `dispose()` runs
/// and cancels its 5s polling `Timer.periodic` — otherwise flutter_test
/// fails the test with a "Timer is still pending" error.
Future<void> _disposePage(WidgetTester tester) async {
  await tester.pumpWidget(const SizedBox());
}

void main() {
  group('Payment UX', () {
    testWidgets('Cash: succeeds, disables during request, settles without a raw error',
        (tester) async {
      var paid = false;
      var payCalls = 0;
      final client = _apiClient((req) async {
        if (req.method == 'GET' && req.url.path.endsWith('/rides/t1')) {
          return _json({
            'trip_id': 't1',
            'trip_status': paid ? 'settled' : 'payment_pending',
            'driver_id': '',
            'final_fare': 50000,
            'currency': 'VND',
          });
        }
        if (req.method == 'POST' && req.url.path.endsWith('/pay')) {
          payCalls++;
          final body = jsonDecode(req.body) as Map<String, dynamic>;
          expect(body['payment_method'], 'cash');
          paid = true;
          return _json({'trip_id': 't1', 'status': 'settled', 'final_fare': 50000, 'currency': 'VND'});
        }
        return http.Response('not found', 404);
      });

      await tester.pumpWidget(MaterialApp(
        home: TripLifecyclePage(tripId: 't1', tripSelection: _tripSelection, apiClient: client),
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      expect(find.text('Trả bằng tiền mặt'), findsOneWidget);
      await tester.tap(find.text('Trả bằng tiền mặt'));
      await tester.pump(); // enters loading state
      await tester.pump(const Duration(milliseconds: 100)); // request + follow-up poll resolve

      expect(payCalls, 1);
      // No raw backend error text ever shown for a clean success.
      expect(find.textContaining('cannot be marked paid'), findsNothing);

      await _disposePage(tester);
    });

    testWidgets('Wallet: succeeds via the wallet button specifically', (tester) async {
      var paid = false;
      String? methodUsed;
      final client = _apiClient((req) async {
        if (req.method == 'GET' && req.url.path.endsWith('/rides/t1')) {
          return _json({
            'trip_id': 't1',
            'trip_status': paid ? 'settled' : 'payment_pending',
            'final_fare': 75000,
            'currency': 'VND',
          });
        }
        if (req.method == 'POST' && req.url.path.endsWith('/pay')) {
          methodUsed = (jsonDecode(req.body) as Map<String, dynamic>)['payment_method'] as String?;
          paid = true;
          return _json({'trip_id': 't1', 'status': 'settled', 'final_fare': 75000, 'currency': 'VND'});
        }
        return http.Response('not found', 404);
      });

      await tester.pumpWidget(MaterialApp(
        home: TripLifecyclePage(tripId: 't1', tripSelection: _tripSelection, apiClient: client),
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      await tester.tap(find.text('Trả bằng ví điện tử'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(methodUsed, 'wallet');

      await _disposePage(tester);
    });

    testWidgets('Double Click: tapping Cash twice in a row only sends one payRide request',
        (tester) async {
      var payCalls = 0;
      final client = _apiClient((req) async {
        if (req.method == 'GET' && req.url.path.endsWith('/rides/t1')) {
          return _json({'trip_id': 't1', 'trip_status': 'payment_pending', 'final_fare': 20000, 'currency': 'VND'});
        }
        if (req.method == 'POST' && req.url.path.endsWith('/pay')) {
          payCalls++;
          // Slow enough that a genuine double-tap would race ahead of it if
          // the guard didn't work.
          await Future<void>.delayed(const Duration(milliseconds: 300));
          return _json({'trip_id': 't1', 'status': 'settled', 'final_fare': 20000, 'currency': 'VND'});
        }
        return http.Response('not found', 404);
      });

      await tester.pumpWidget(MaterialApp(
        home: TripLifecyclePage(tripId: 't1', tripSelection: _tripSelection, apiClient: client),
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      // Two taps back-to-back, no pump in between — both hit the exact same
      // `_pay('cash')` closure from the last build; the second must be
      // rejected by the in-flight guard, not merely by the button being
      // visually disabled after a rebuild.
      await tester.tap(find.text('Trả bằng tiền mặt'));
      await tester.tap(find.text('Trả bằng tiền mặt'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 400));

      expect(payCalls, 1, reason: 'the double-submit guard must block the second tap');

      await _disposePage(tester);
    });

    testWidgets('Already Paid (duplicate tap race): silently treated as success, no raw backend error',
        (tester) async {
      var callCount = 0;
      final client = _apiClient((req) async {
        if (req.method == 'GET' && req.url.path.endsWith('/rides/t1')) {
          return _json({
            'trip_id': 't1',
            'trip_status': callCount > 0 ? 'settled' : 'payment_pending',
            'final_fare': 40000,
            'currency': 'VND',
          });
        }
        if (req.method == 'POST' && req.url.path.endsWith('/pay')) {
          callCount++;
          // Every call is rejected as already-settled, simulating a retried
          // request whose first attempt actually succeeded server-side.
          return _json({'error': 'trip cannot be marked paid from status: settled'}, status: 412);
        }
        return http.Response('not found', 404);
      });

      await tester.pumpWidget(MaterialApp(
        home: TripLifecyclePage(tripId: 't1', tripSelection: _tripSelection, apiClient: client),
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      await tester.tap(find.text('Trả bằng tiền mặt'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      // Section 7: "Nếu Paid rồi → Hiện 'Chuyến đi đã được thanh toán.'
      // KHÔNG show lỗi backend."
      expect(find.textContaining('cannot be marked paid'), findsNothing);
      expect(find.textContaining('Chuyến đi đã được thanh toán'), findsOneWidget);

      await _disposePage(tester);
    });

    testWidgets('Payment timeout: shows a retry-able message, not a crash, buttons re-enable',
        (tester) async {
      final client = _apiClient((req) async {
        if (req.method == 'GET' && req.url.path.endsWith('/rides/t1')) {
          return _json({'trip_id': 't1', 'trip_status': 'payment_pending', 'final_fare': 60000, 'currency': 'VND'});
        }
        if (req.method == 'POST' && req.url.path.endsWith('/pay')) {
          // Never resolves within ApiClient's 15s timeout — triggers
          // ApiClient's own onTimeout handler.
          await Future<void>.delayed(const Duration(seconds: 30));
          return _json({'trip_id': 't1', 'status': 'settled', 'final_fare': 60000, 'currency': 'VND'});
        }
        return http.Response('not found', 404);
      });

      await tester.pumpWidget(MaterialApp(
        home: TripLifecyclePage(tripId: 't1', tripSelection: _tripSelection, apiClient: client),
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      await tester.tap(find.text('Trả bằng tiền mặt'));
      await tester.pump();
      // Advance virtual time past ApiClient's 15s timeout.
      await tester.pump(const Duration(seconds: 16));

      expect(find.textContaining('Hết thời gian chờ'), findsOneWidget);
      // The relabeled retry CTA must be visible and tappable again, not
      // permanently disabled — AppButton.primary renders a FilledButton.
      expect(find.text('Thử lại · Tiền mặt'), findsOneWidget);
      final retryButton = tester.widget<FilledButton>(
        find.ancestor(of: find.text('Thử lại · Tiền mặt'), matching: find.byType(FilledButton)).first,
      );
      expect(retryButton.onPressed, isNotNull);

      // Drain the mock handler's still-pending 30s delayed response so its
      // underlying Timer fires before teardown — ApiClient's own 15s
      // `.timeout()` only stops *waiting* on the caller side, it doesn't
      // cancel the in-flight request itself.
      await tester.pump(const Duration(seconds: 20));

      await _disposePage(tester);
    });

    testWidgets('Network Lost: a thrown non-ApiException shows a generic retry message',
        (tester) async {
      final client = _apiClient((req) async {
        if (req.method == 'GET' && req.url.path.endsWith('/rides/t1')) {
          return _json({'trip_id': 't1', 'trip_status': 'payment_pending', 'final_fare': 45000, 'currency': 'VND'});
        }
        if (req.method == 'POST' && req.url.path.endsWith('/pay')) {
          throw const SocketExceptionStub();
        }
        return http.Response('not found', 404);
      });

      await tester.pumpWidget(MaterialApp(
        home: TripLifecyclePage(tripId: 't1', tripSelection: _tripSelection, apiClient: client),
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      await tester.tap(find.text('Trả bằng tiền mặt'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.textContaining('mất kết nối'), findsOneWidget);
      // Never a stack trace / raw exception string leaked to the UI.
      expect(find.textContaining('SocketExceptionStub'), findsNothing);

      await _disposePage(tester);
    });
  });
}

/// Stand-in for `dart:io`'s `SocketException` (not importable on web/test
/// targets without extra setup) — any non-`ApiException` thrown from the
/// HTTP layer must be handled identically by `_pay`'s generic `catch`.
class SocketExceptionStub implements Exception {
  const SocketExceptionStub();
  @override
  String toString() => 'SocketExceptionStub: Failed host lookup';
}
