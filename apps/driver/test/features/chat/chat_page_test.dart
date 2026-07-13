// Widget test for the Driver side of the Communication Module's in-app chat
// (Part 2). Uses ApiClient's injectable `http.Client` so this runs against
// a `MockClient` — no real backend needed.
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:driver/core/auth/auth_state.dart';
import 'package:driver/core/network/api_client.dart';
import 'package:driver/features/chat/presentation/pages/chat_page.dart';

http.Response _json(Map<String, dynamic> body, {int status = 200}) => http.Response(
      jsonEncode(body),
      status,
      headers: {'content-type': 'application/json'},
    );

void main() {
  testWidgets('ChatPage loads the conversation and lists existing messages', (tester) async {
    var pollCount = 0;
    final client = ApiClient(
      baseUrl: 'http://test.local',
      authState: AuthState(),
      httpClient: MockClient((req) async {
        if (req.url.path.endsWith('/conversation')) {
          return _json({
            'id': 'conv1',
            'trip_id': 't1',
            'rider_id': 'rider1',
            'driver_id': 'driver1',
            'trip_type': 'ride',
            'status': 'open',
            'unread_count': 0,
          });
        }
        if (req.url.path.contains('/messages')) {
          final isPoll = req.url.queryParameters['poll'] == 'true';
          if (isPoll) {
            pollCount++;
            // Real long-poll would hold this open server-side; the mock
            // resolves immediately with no new messages so the test doesn't
            // need to wait out a real timeout — ChatPage's own 400ms
            // re-poll backoff is what's actually being exercised here.
            return _json({'messages': []});
          }
          return _json({
            'messages': [
              {
                'id': 'm1',
                'seq': 1,
                'conversation_id': 'conv1',
                'sender_id': 'rider1',
                'sender_role': 'rider',
                'body': 'Xin chào, tôi đang chờ ở cổng',
                'quick_reply_key': '',
                'created_at': DateTime.now().toUtc().toIso8601String(),
              },
            ],
          });
        }
        return _json({});
      }),
    );

    await tester.pumpWidget(MaterialApp(home: ChatPage(tripId: 't1', apiClient: client)));
    await tester.pump();
    await tester.pump();
    await tester.pump();

    expect(find.text('Xin chào, tôi đang chờ ở cổng'), findsOneWidget);
    expect(pollCount, greaterThan(0));

    // Dispose the page so its 400ms re-poll Timer chain is cancelled —
    // otherwise flutter_test fails with "A Timer is still pending" since
    // ChatPage's long-poll loop re-schedules itself indefinitely.
    await tester.pumpWidget(const SizedBox());
  });
}
