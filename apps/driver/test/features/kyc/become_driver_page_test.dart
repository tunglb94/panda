// Regression test for the KYC wizard's Step 1 "Tiếp tục" button: it must
// re-evaluate (and enable) after the rider types into the personal-info
// fields, not only when some unrelated widget (e.g. the date-of-birth
// picker) happens to trigger a rebuild. Before the fix, typing into
// Họ và tên/Địa chỉ/Số CCCD after already picking a date of birth left the
// button permanently disabled — exactly the "filled everything in but
// can't press Next" bug report.
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

import 'package:driver/core/auth/auth_state.dart';
import 'package:driver/core/network/api_client.dart';
import 'package:driver/features/kyc/presentation/pages/become_driver_page.dart';

http.Response _json(Map<String, dynamic> body, {int status = 200}) => http.Response(
      jsonEncode(body),
      status,
      headers: {'content-type': 'application/json'},
    );

ApiClient _mockApiClient() => ApiClient(
      baseUrl: 'http://test.local',
      authState: AuthState(),
      httpClient: MockClient((req) async {
        if (req.url.path.endsWith('/documents')) {
          return _json({'documents': []});
        }
        // Nothing submitted yet — both verification GETs 404.
        return _json({'error': 'not found'}, status: 404);
      }),
    );

FilledButton _findContinueButton(WidgetTester tester) =>
    tester.widget<FilledButton>(find.byWidgetPredicate((w) => w is FilledButton));

void main() {
  testWidgets('Step 1 "Tiếp tục" enables after typing, even when typing happens after picking the date of birth',
      (tester) async {
    await tester.pumpWidget(MaterialApp(home: BecomeDriverPage(apiClient: _mockApiClient())));
    await tester.pumpAndSettle();

    // Nothing filled in yet — button must be disabled.
    expect(_findContinueButton(tester).onPressed, isNull);

    // Pick the date of birth FIRST (this is the rebuild that, before the
    // fix, was the ONLY thing keeping the button in sync with validity).
    await tester.tap(find.text('Chọn ngày sinh'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('OK'));
    await tester.pumpAndSettle();

    // Still disabled — the 3 text fields are still empty.
    expect(_findContinueButton(tester).onPressed, isNull);

    // Now type into the fields AFTER the date-of-birth rebuild — this is
    // exactly the order that was broken.
    await tester.enterText(find.widgetWithText(TextField, 'Họ và tên'), 'Nguyễn Văn A');
    await tester.enterText(find.widgetWithText(TextField, 'Địa chỉ thường trú'), '123 Lê Lợi, Q1');
    await tester.enterText(find.widgetWithText(TextField, 'Số CCCD'), '079095001234');
    await tester.pump();

    expect(_findContinueButton(tester).onPressed, isNotNull,
        reason: 'all 4 required fields are filled — Tiếp tục must be enabled');
  });
}
