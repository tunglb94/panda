import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:admin/core/auth/auth_state.dart';
import 'package:admin/core/network/api_client.dart';
import 'package:admin/core/storage/token_storage.dart';
import 'package:admin/features/auth/presentation/pages/login_page.dart';

void main() {
  testWidgets('LoginPage shows the phone field and login button', (tester) async {
    final authState = AuthState();
    final tokenStorage = TokenStorage();
    final apiClient = ApiClient(baseUrl: 'https://example.invalid', authState: authState);

    await tester.pumpWidget(MaterialApp(
      home: LoginPage(authState: authState, tokenStorage: tokenStorage, apiClient: apiClient),
    ));

    expect(find.text('Panda Admin'), findsOneWidget);
    expect(find.widgetWithText(TextField, 'Số điện thoại Admin'), findsOneWidget);
    expect(find.widgetWithText(FilledButton, 'Đăng nhập'), findsOneWidget);
  });
}
