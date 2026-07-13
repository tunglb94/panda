// Widget tests for the Driver Finance Wallet screen (Phần 13). Uses
// ApiClient's injectable `http.Client` (mirrors the KYC/Communication
// Module tests) so these run against a `MockClient` — no real backend
// needed. SharedPreferences is mocked in-memory per Flutter's own testing
// convention (Phần 11 — Offline cache).
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:driver/core/auth/auth_state.dart';
import 'package:driver/core/network/api_client.dart';
import 'package:driver/features/wallet/presentation/pages/wallet_page.dart';
import 'package:driver/features/wallet/presentation/widgets/wallet_transaction_tile.dart';

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

Map<String, dynamic> _summaryJson({
  int available = 800000,
  int pending = 0,
  int outstanding = 0,
}) =>
    {
      'currency': 'VND',
      'available_cents': available,
      'pending_cents': pending,
      'outstanding_cents': outstanding,
      'net_cents': available - outstanding,
      'lifetime_earned_cents': available,
      'lifetime_withdrawn_cents': 0,
    };

void main() {
  setUp(() {
    SharedPreferences.setMockInitialValues({});
  });

  testWidgets('WalletPage shows Available balance and Rút tiền button', (tester) async {
    final client = _mockApiClient((req) async {
      final path = req.url.path;
      if (path.contains('/wallet/summary')) return _json(_summaryJson());
      if (path.contains('/wallet/transactions')) return _json({'transactions': []});
      if (path.contains('/wallet/bank-account')) return _json({'error': 'not found'}, status: 404);
      if (path.contains('/wallet/payouts')) return _json({'payout_requests': []});
      return _json({}, status: 404);
    });

    await tester.pumpWidget(MaterialApp(home: WalletPage(apiClient: client)));
    await tester.pumpAndSettle();

    expect(find.text('800.000 đ'), findsOneWidget);
    expect(find.text('Rút tiền'), findsOneWidget);
  });

  testWidgets('WalletPage disables Rút tiền when there is no bank account', (tester) async {
    final client = _mockApiClient((req) async {
      final path = req.url.path;
      if (path.contains('/wallet/summary')) return _json(_summaryJson());
      if (path.contains('/wallet/transactions')) return _json({'transactions': []});
      if (path.contains('/wallet/bank-account')) return _json({'error': 'not found'}, status: 404);
      if (path.contains('/wallet/payouts')) return _json({'payout_requests': []});
      return _json({}, status: 404);
    });

    await tester.pumpWidget(MaterialApp(home: WalletPage(apiClient: client)));
    await tester.pumpAndSettle();

    final button = tester.widget<FilledButton>(find.byWidgetPredicate((w) => w is FilledButton));
    expect(button.onPressed, isNull);
  });

  testWidgets('WalletPage enables Rút tiền when eligible', (tester) async {
    final client = _mockApiClient((req) async {
      final path = req.url.path;
      if (path.contains('/wallet/summary')) return _json(_summaryJson());
      if (path.contains('/wallet/transactions')) return _json({'transactions': []});
      if (path.contains('/wallet/bank-account')) {
        return _json({
          'bank_name': 'Vietcombank',
          'account_holder_name': 'Nguyen Van A',
          'masked_account_number': '••••6789',
          'branch_name': '',
          'updated_at': '2026-07-01T00:00:00Z',
        });
      }
      if (path.contains('/wallet/payouts')) return _json({'payout_requests': []});
      return _json({}, status: 404);
    });

    await tester.pumpWidget(MaterialApp(home: WalletPage(apiClient: client)));
    await tester.pumpAndSettle();

    final button = tester.widget<FilledButton>(find.byWidgetPredicate((w) => w is FilledButton));
    expect(button.onPressed, isNotNull);
    expect(find.textContaining('Vietcombank'), findsOneWidget);
  });

  testWidgets('WalletPage shows the Outstanding warning banner', (tester) async {
    final client = _mockApiClient((req) async {
      final path = req.url.path;
      if (path.contains('/wallet/summary')) return _json(_summaryJson(outstanding: 24000));
      if (path.contains('/wallet/transactions')) return _json({'transactions': []});
      if (path.contains('/wallet/bank-account')) return _json({'error': 'not found'}, status: 404);
      if (path.contains('/wallet/payouts')) return _json({'payout_requests': []});
      return _json({}, status: 404);
    });

    await tester.pumpWidget(MaterialApp(home: WalletPage(apiClient: client)));
    await tester.pumpAndSettle();

    expect(find.textContaining('nợ Panda'), findsOneWidget);
  });

  testWidgets('WalletPage renders transaction history rows', (tester) async {
    final client = _mockApiClient((req) async {
      final path = req.url.path;
      if (path.contains('/wallet/summary')) return _json(_summaryJson());
      if (path.contains('/wallet/transactions')) {
        return _json({
          'transactions': [
            {
              'type': 'ride_income',
              'direction': 'credit',
              'amount_cents': 96000,
              'currency': 'VND',
              'description': 'Thu nhập chuyến xe trip-1',
              'payment_method': 'wallet',
              'created_at': '2026-07-10T08:00:00Z',
            },
          ],
        });
      }
      if (path.contains('/wallet/bank-account')) return _json({'error': 'not found'}, status: 404);
      if (path.contains('/wallet/payouts')) return _json({'payout_requests': []});
      return _json({}, status: 404);
    });

    await tester.pumpWidget(MaterialApp(home: WalletPage(apiClient: client)));
    await tester.pumpAndSettle();

    await tester.dragUntilVisible(
      find.byType(WalletTransactionTile),
      find.byType(ListView),
      const Offset(0, -300),
    );

    expect(find.byType(WalletTransactionTile), findsOneWidget);
  });
}
