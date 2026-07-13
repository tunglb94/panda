// Widget tests for the redesigned Home hub — Bike/Car quick-pick cards with
// real vehicle artwork, Gửi hàng/Đặt đồ ăn cards, and the bottom banner.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';

import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/routing/google_route_service.dart';
import 'package:rider/features/home/presentation/pages/home_hub_page.dart';

ApiClient _testApiClient() => ApiClient(baseUrl: 'http://test.local', authState: AuthState());

void main() {
  Widget wrap(Widget child) => MaterialApp.router(
        routerConfig: GoRouter(
          initialLocation: '/',
          routes: [
            GoRoute(path: '/', builder: (context, state) => child),
            GoRoute(path: '/wallet', builder: (context, state) => const SizedBox()),
            GoRoute(path: '/profile', builder: (context, state) => const SizedBox()),
          ],
        ),
      );

  testWidgets('HomeHubPage shows Bike/Car quick picks, Gửi hàng/Đặt đồ ăn, and the banner', (tester) async {
    await tester.pumpWidget(wrap(HomeHubPage(
      apiClient: _testApiClient(),
      routeProvider: const GoogleRouteProvider(apiKey: ''),
    )));
    await tester.pumpAndSettle();

    expect(find.text('Bike'), findsOneWidget);
    expect(find.text('Car'), findsOneWidget);
    expect(find.text('Gửi hàng'), findsOneWidget);
    expect(find.text('Đặt đồ ăn'), findsOneWidget);
    // 4 primary card images (Bike, Car, Gửi hàng, Đặt đồ ăn) + the banner.
    expect(find.byType(Image), findsNWidgets(5));
  });

  testWidgets('Tapping Đặt đồ ăn shows the coming-soon message, not a real flow', (tester) async {
    await tester.pumpWidget(wrap(HomeHubPage(
      apiClient: _testApiClient(),
      routeProvider: const GoogleRouteProvider(apiKey: ''),
    )));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Đặt đồ ăn'));
    await tester.pump();

    expect(find.textContaining('chưa khả dụng'), findsOneWidget);
  });
}
