// Basic smoke test for the Booking UI module (Phase R-01).
//
// The previous version of this file was the unmodified `flutter create`
// counter-app template and referenced a `MyApp` class that has never existed
// in this codebase (the real root widget is `RiderApp`, see lib/app.dart) —
// it only ever failed static analysis, it was never actually run as a
// regression test. This replaces it with a minimal render check for the new
// Booking module using mock data, with no platform-channel dependencies
// (no GoogleMap widget is instantiated here, so no plugin mocking is
// required).

import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/storage/token_storage.dart';
import 'package:rider/features/booking/presentation/pages/booking_page.dart';
import 'package:rider/features/history/presentation/pages/trip_history_page.dart';
import 'package:rider/features/profile/presentation/pages/notification_center_page.dart';
import 'package:rider/features/profile/presentation/pages/profile_page.dart';
import 'package:rider/features/profile/presentation/pages/settings_page.dart';
import 'package:rider/features/trip/domain/models/rider_trip_status.dart';
import 'package:rider/features/trip/presentation/pages/trip_preview_menu_page.dart';
import 'package:rider/features/trip/presentation/pages/trip_state_preview_page.dart';

/// Real (not mocked) dependencies for widgets that now require them —
/// no network call is actually made unless a test explicitly settles past
/// one, since none of the tests below assert on live backend data.
ApiClient _testApiClient() =>
    ApiClient(baseUrl: 'http://localhost:8080', authState: AuthState());

http.Response _json(Map<String, dynamic> body, {int status = 200}) => http.Response(
      jsonEncode(body),
      status,
      headers: {'content-type': 'application/json'},
    );

ApiClient _mockApiClient(Future<http.Response> Function(http.Request) handler) => ApiClient(
      baseUrl: 'http://test.local',
      authState: AuthState(),
      httpClient: MockClient((req) => handler(req)),
    );

void main() {
  testWidgets('BookingPage renders trip summary and vehicle options',
      (WidgetTester tester) async {
    await tester.pumpWidget(MaterialApp(home: BookingPage(apiClient: _testApiClient())));

    expect(find.text('Book a Ride'), findsOneWidget);
    expect(find.text('Choose a ride'), findsOneWidget);
    expect(find.text('Car'), findsOneWidget);
    expect(find.text('Moto'), findsOneWidget);
    expect(find.text('Van'), findsOneWidget);
  });

  testWidgets('TripPreviewMenuPage lists all five trip lifecycle states',
      (WidgetTester tester) async {
    await tester.pumpWidget(MaterialApp(home: TripPreviewMenuPage(apiClient: _testApiClient())));

    for (final status in RiderTripStatus.values) {
      expect(find.text(status.label), findsOneWidget);
    }
  });

  testWidgets(
      'TripStatePreviewPage renders Driver Assigned state with driver info',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: TripStatePreviewPage(
          status: RiderTripStatus.driverAssigned,
          apiClient: _testApiClient(),
        ),
      ),
    );

    expect(find.text('Driver Assigned'), findsWidgets);
    expect(find.text('Nguyen Van A'), findsOneWidget);
    expect(find.text('Cancel Ride'), findsOneWidget);
    expect(find.text('Emergency'), findsOneWidget);
  });

  testWidgets('TripStatePreviewPage renders Trip Completed state with fare',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: TripStatePreviewPage(status: RiderTripStatus.completed, apiClient: _testApiClient()),
      ),
    );

    expect(find.text('Fare summary'), findsOneWidget);
    expect(find.widgetWithText(FilledButton, 'Done'), findsOneWidget);
    expect(find.text('Cancel Ride'), findsNothing);
  });

  testWidgets('ProfilePage shows loading then mock profile info',
      (WidgetTester tester) async {
    await tester.pumpWidget(MaterialApp(
      home: ProfilePage(
        authState: AuthState(),
        tokenStorage: TokenStorage(),
        apiClient: _testApiClient(),
      ),
    ));

    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    await tester.pumpAndSettle();

    expect(find.text('Alex Rider'), findsOneWidget);
    expect(find.text('Gold Member'), findsOneWidget);
    expect(find.text('Settings'), findsOneWidget);
  });

  testWidgets('SettingsPage lists all required settings entries',
      (WidgetTester tester) async {
    // The settings list is taller than the default test viewport, and
    // ListView virtualises off-screen children even with a plain
    // `children:` list — grow the surface so every section is actually
    // built, rather than asserting on a real scrolling interaction.
    tester.view.physicalSize = const Size(400, 1600);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.reset);

    await tester.pumpWidget(MaterialApp(home: SettingsPage(apiClient: _testApiClient())));

    for (final label in [
      'Personal Information',
      'Payment Methods',
      'Notifications',
      'Privacy',
      'Security',
      'Language',
      'Help Center',
      'About',
      'Logout',
    ]) {
      expect(find.text(label), findsOneWidget);
    }
  });

  testWidgets('NotificationCenterPage shows real notifications from the API',
      (WidgetTester tester) async {
    final client = _mockApiClient((req) async {
      return _json({
        'notifications': [
          {
            'id': 'n1',
            'category': 'trip',
            'title': 'Chuyến đi đã hoàn tất',
            'body': 'Cảm ơn bạn đã sử dụng Panda.',
            'trip_id': 't1',
            'conversation_id': '',
            'created_at': DateTime.now().toUtc().toIso8601String(),
            'is_read': false,
          },
          {
            'id': 'n2',
            'category': 'chat',
            'title': 'Tin nhắn mới',
            'body': 'Tôi tới rồi',
            'trip_id': 't1',
            'conversation_id': 'c1',
            'created_at': DateTime.now().toUtc().toIso8601String(),
            'is_read': true,
          },
        ],
        'unread_count': 1,
      });
    });
    await tester.pumpWidget(MaterialApp(home: NotificationCenterPage(apiClient: client)));
    await tester.pumpAndSettle();

    expect(find.text('Chuyến đi đã hoàn tất'), findsOneWidget);
    expect(find.text('Tin nhắn mới'), findsOneWidget);
  });

  testWidgets('NotificationCenterPage shows empty state when the feed is empty',
      (WidgetTester tester) async {
    final client = _mockApiClient((req) async => _json({'notifications': [], 'unread_count': 0}));
    await tester.pumpWidget(MaterialApp(home: NotificationCenterPage(apiClient: client)));
    await tester.pumpAndSettle();

    expect(find.text('Chưa có thông báo nào'), findsOneWidget);
  });

  testWidgets('NotificationCenterPage shows error state on a backend failure',
      (WidgetTester tester) async {
    final client = _mockApiClient((req) async => _json({'error': 'internal'}, status: 500));
    await tester.pumpWidget(MaterialApp(home: NotificationCenterPage(apiClient: client)));
    await tester.pumpAndSettle();

    expect(find.text('Không thể tải thông báo'), findsOneWidget);
  });

  // NOTE (Closed Beta polish pass): the previous TripHistoryPage/
  // TripDetailPage/ReceiptPage tests here asserted on an English-language,
  // mock-catalog-driven, filtered/grouped history flow that no longer
  // matches the real app — the live `TripHistoryPage` has always fetched
  // `GET /api/v1/rider/trips` directly (Vietnamese strings, no filter
  // chips, no dev preview menu), and `TripDetailPage`/`ReceiptPage` were
  // dead code (never reachable from any route) built against a richer mock
  // model than the backend actually returns. Both were removed as part of
  // the design-system/cleanup pass; `TripHistoryPage` was rebuilt on the
  // shared component library with real tap-through navigation to a new,
  // honest `TripDetailPage`. Fresh tests for that flow are a follow-up —
  // see the sprint report's P1 backlog.
  testWidgets('TripHistoryPage shows loading then the real trip list',
      (WidgetTester tester) async {
    await tester.pumpWidget(MaterialApp(home: TripHistoryPage(apiClient: _testApiClient())));

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
  });
}
