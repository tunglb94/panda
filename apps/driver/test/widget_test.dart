// Smoke tests for the Driver app shell (Phase D-01).
//
// The default `flutter create` counter-app template referenced a `MyApp`
// class that no longer exists (the real root widget is `DriverApp`, see
// lib/app.dart) — replaced with real checks of the 5-tab shell and the
// Developer page, consistent with how `apps/rider` fixed the same
// leftover template in Phase R-01.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:driver/app.dart';
import 'package:driver/features/home/presentation/pages/home_page.dart';
import 'package:driver/features/profile/presentation/pages/developer_page.dart';

void main() {
  testWidgets('DriverApp shows Home tab and all 5 bottom nav destinations',
      (WidgetTester tester) async {
    await tester.pumpWidget(const DriverApp());
    await tester.pumpAndSettle();

    expect(find.text('Home'), findsWidgets);
    expect(find.widgetWithText(NavigationDestination, 'Trips'), findsOneWidget);
    // "Earnings" now also appears as a Home dashboard Quick Action label
    // (Phase D-02), so assert on the nav destination specifically.
    expect(find.widgetWithText(NavigationDestination, 'Earnings'), findsOneWidget);
    expect(find.widgetWithText(NavigationDestination, 'Notifications'), findsOneWidget);
    expect(find.widgetWithText(NavigationDestination, 'Profile'), findsOneWidget);
  });

  testWidgets('Tapping a bottom nav destination switches tabs',
      (WidgetTester tester) async {
    await tester.pumpWidget(const DriverApp());
    await tester.pumpAndSettle();

    await tester.tap(find.widgetWithText(NavigationDestination, 'Earnings'));
    await tester.pumpAndSettle();

    expect(
      find.descendant(
        of: find.byType(AppBar),
        matching: find.text('Earnings'),
      ),
      findsOneWidget,
    );
  });

  testWidgets('Profile tab Developer entry opens the Developer page',
      (WidgetTester tester) async {
    await tester.pumpWidget(const DriverApp());
    await tester.pumpAndSettle();

    await tester.tap(find.widgetWithText(NavigationDestination, 'Profile'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(ListTile, 'Developer'));
    await tester.pumpAndSettle();

    expect(find.byType(DeveloperPage), findsOneWidget);
    expect(find.text('App version'), findsOneWidget);
    expect(find.text('Build mode'), findsOneWidget);
    expect(find.text('Debug'), findsOneWidget);
    expect(find.text('Flutter version'), findsOneWidget);
    expect(find.text('Environment'), findsOneWidget);
  });

  testWidgets('HomePage shows loading then driver summary and stats',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: HomePage()));

    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    await tester.pumpAndSettle();

    expect(find.text('Nguyen Van A'), findsOneWidget);
    expect(find.text('Toyota Vios · 51G-123.45'), findsOneWidget);
    expect(find.text('7'), findsOneWidget);
    expect(find.text("You're Offline — tap to go online"), findsOneWidget);
    expect(find.text('Quick actions'), findsOneWidget);
  });

  testWidgets(
      'Availability toggle walks through all 4 states and updates the status card',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: HomePage()));
    await tester.pumpAndSettle();

    expect(find.text("You're offline"), findsOneWidget);

    await tester.tap(find.text("You're Offline — tap to go online"));
    await tester.pump();
    expect(find.text('Going online…'), findsOneWidget);

    await tester.pump(const Duration(milliseconds: 1300));
    expect(find.text("You're Online — tap to go offline"), findsOneWidget);
    expect(find.text('Waiting for trips'), findsOneWidget);

    await tester.pump(const Duration(seconds: 4));
    expect(find.text('Searching nearby'), findsOneWidget);

    await tester.tap(find.text("You're Online — tap to go offline"));
    await tester.pump();
    expect(find.text('Going offline…'), findsOneWidget);

    await tester.pump(const Duration(milliseconds: 1000));
    expect(find.text("You're Offline — tap to go online"), findsOneWidget);
  });

  testWidgets('HomePage shows empty state (no vehicle) via dev preview menu',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: HomePage()));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.tune));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Empty (dev)'));
    await tester.pumpAndSettle();

    expect(find.text('No vehicle assigned yet'), findsOneWidget);
  });

  testWidgets('HomePage shows error state via dev preview menu',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: HomePage()));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.tune));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Error (dev)'));
    await tester.pumpAndSettle();

    expect(find.text("Couldn't load your dashboard"), findsOneWidget);
  });

  testWidgets('Quick action tap shows a placeholder message',
      (WidgetTester tester) async {
    // The dashboard (header + stats + toggle + status + 4 quick actions) is
    // taller than the default test viewport — grow it so the Quick Actions
    // grid is actually laid out on-screen, rather than asserting on a real
    // scroll gesture (same fix used for Rider's SettingsPage test, R-03).
    tester.view.physicalSize = const Size(400, 1400);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.reset);

    await tester.pumpWidget(const MaterialApp(home: HomePage()));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Support'));
    await tester.pump();

    expect(
      find.text('Support is a placeholder — not yet implemented.'),
      findsOneWidget,
    );
  });
}
