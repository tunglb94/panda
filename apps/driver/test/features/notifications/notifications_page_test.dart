// Widget tests for the Communication Module's real (non-mock) Notification
// Center on the Driver side (Part 3). Uses ApiClient's injectable
// `http.Client` (mirrors apps/rider) so these run against a `MockClient` —
// no real backend needed.
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:driver/core/auth/auth_state.dart';
import 'package:driver/core/network/api_client.dart';
import 'package:driver/features/notifications/presentation/pages/notifications_page.dart';

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
  testWidgets('NotificationsPage shows real notifications from the API', (tester) async {
    final client = _mockApiClient((req) async {
      return _json({
        'notifications': [
          {
            'id': 'n1',
            'category': 'trip',
            'title': 'Tài xế đã nhận chuyến',
            'body': 'Tài xế đang trên đường đến điểm đón.',
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

    await tester.pumpWidget(MaterialApp(home: NotificationsPage(apiClient: client)));
    await tester.pump();
    await tester.pump();

    expect(find.text('Tài xế đã nhận chuyến'), findsOneWidget);
    expect(find.text('Tin nhắn mới'), findsOneWidget);
  });

  testWidgets('NotificationsPage shows empty state when the feed is empty', (tester) async {
    final client = _mockApiClient((req) async => _json({'notifications': [], 'unread_count': 0}));

    await tester.pumpWidget(MaterialApp(home: NotificationsPage(apiClient: client)));
    await tester.pump();
    await tester.pump();

    expect(find.text('Không có thông báo'), findsOneWidget);
  });

  testWidgets('NotificationsPage shows an error state on a backend failure', (tester) async {
    final client = _mockApiClient((req) async => _json({'error': 'internal'}, status: 500));

    await tester.pumpWidget(MaterialApp(home: NotificationsPage(apiClient: client)));
    await tester.pump();
    await tester.pump();

    expect(find.textContaining('Không thể tải thông báo'), findsOneWidget);
  });
}
