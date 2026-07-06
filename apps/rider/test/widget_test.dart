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

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:rider/features/booking/presentation/pages/booking_page.dart';
import 'package:rider/features/history/domain/models/mock_trip_history_catalog.dart';
import 'package:rider/features/history/presentation/pages/receipt_page.dart';
import 'package:rider/features/history/presentation/pages/trip_detail_page.dart';
import 'package:rider/features/history/presentation/pages/trip_history_page.dart';
import 'package:rider/features/profile/presentation/pages/notification_center_page.dart';
import 'package:rider/features/profile/presentation/pages/profile_page.dart';
import 'package:rider/features/profile/presentation/pages/settings_page.dart';
import 'package:rider/features/trip/domain/models/rider_trip_status.dart';
import 'package:rider/features/trip/presentation/pages/trip_preview_menu_page.dart';
import 'package:rider/features/trip/presentation/pages/trip_state_preview_page.dart';

void main() {
  testWidgets('BookingPage renders trip summary and vehicle options',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: BookingPage()));

    expect(find.text('Book a Ride'), findsOneWidget);
    expect(find.text('Choose a ride'), findsOneWidget);
    expect(find.text('Car'), findsOneWidget);
    expect(find.text('Moto'), findsOneWidget);
    expect(find.text('Van'), findsOneWidget);
  });

  testWidgets('TripPreviewMenuPage lists all five trip lifecycle states',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripPreviewMenuPage()));

    for (final status in RiderTripStatus.values) {
      expect(find.text(status.label), findsOneWidget);
    }
  });

  testWidgets(
      'TripStatePreviewPage renders Driver Assigned state with driver info',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: TripStatePreviewPage(status: RiderTripStatus.driverAssigned),
      ),
    );

    expect(find.text('Driver Assigned'), findsWidgets);
    expect(find.text('Nguyen Van A'), findsOneWidget);
    expect(find.text('Cancel Ride'), findsOneWidget);
    expect(find.text('Contact Driver'), findsOneWidget);
    expect(find.text('Emergency'), findsOneWidget);
  });

  testWidgets('TripStatePreviewPage renders Trip Completed state with fare',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: TripStatePreviewPage(status: RiderTripStatus.completed),
      ),
    );

    expect(find.text('Fare summary'), findsOneWidget);
    expect(find.widgetWithText(FilledButton, 'Done'), findsOneWidget);
    expect(find.text('Cancel Ride'), findsNothing);
  });

  testWidgets('ProfilePage shows loading then mock profile info',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: ProfilePage()));

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

    await tester.pumpWidget(const MaterialApp(home: SettingsPage()));

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

  testWidgets('NotificationCenterPage shows mock notifications',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: NotificationCenterPage()));
    await tester.pumpAndSettle();

    expect(find.text('Trip completed'), findsOneWidget);
    expect(find.text('Welcome to FAIRRIDE'), findsOneWidget);
  });

  testWidgets('NotificationCenterPage shows empty state via dev preview menu',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: NotificationCenterPage()));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.tune));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Empty (dev)'));
    await tester.pumpAndSettle();

    expect(find.text('No notifications yet'), findsOneWidget);
  });

  testWidgets('NotificationCenterPage shows error state via dev preview menu',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: NotificationCenterPage()));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.tune));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Error (dev)'));
    await tester.pumpAndSettle();

    expect(find.text("Couldn't load notifications"), findsOneWidget);
  });

  testWidgets('TripHistoryPage shows loading then grouped mock trips',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(400, 1800);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.reset);

    await tester.pumpWidget(const MaterialApp(home: TripHistoryPage()));

    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    await tester.pumpAndSettle();

    // "Today" legitimately appears twice: the date-group header and the
    // "Today" date-filter chip.
    expect(find.text('Today'), findsWidgets);
    expect(find.text('Nguyen Van A · Toyota Vios'), findsOneWidget);
  });

  testWidgets('TripHistoryPage Completed filter hides cancelled trips',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(400, 1800);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.reset);

    await tester.pumpWidget(const MaterialApp(home: TripHistoryPage()));
    await tester.pumpAndSettle();

    expect(find.text('Le Van C · Ford Transit'), findsOneWidget);

    await tester.tap(find.widgetWithText(ChoiceChip, 'Completed'));
    await tester.pumpAndSettle();

    expect(find.text('Le Van C · Ford Transit'), findsNothing);
  });

  testWidgets('TripHistoryPage shows empty state via dev preview menu',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripHistoryPage()));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.tune));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Empty (dev)'));
    await tester.pumpAndSettle();

    expect(find.text('No trips yet'), findsOneWidget);
  });

  testWidgets('TripHistoryPage shows error state via dev preview menu',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripHistoryPage()));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.tune));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Error (dev)'));
    await tester.pumpAndSettle();

    expect(find.text("Couldn't load trip history"), findsOneWidget);
  });

  testWidgets('TripDetailPage shows route, driver, timeline, and fare',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(400, 2000);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(tester.view.reset);

    final entry = MockTripHistoryCatalog.sample().first;
    await tester.pumpWidget(MaterialApp(home: TripDetailPage(entry: entry)));

    expect(find.text('Route summary'), findsOneWidget);
    expect(find.text('Nguyen Van A'), findsOneWidget);
    expect(find.text('Timeline'), findsOneWidget);
    expect(find.text('Fare summary'), findsOneWidget);
    expect(find.text('View Receipt'), findsOneWidget);
  });

  testWidgets('ReceiptPage shows trip id, rider, and total',
      (WidgetTester tester) async {
    final entry = MockTripHistoryCatalog.sample().first;
    await tester.pumpWidget(MaterialApp(home: ReceiptPage(entry: entry)));

    expect(find.text('Trip Receipt'), findsOneWidget);
    expect(find.text(entry.id), findsOneWidget);
    expect(find.text('Alex Rider'), findsOneWidget);
    expect(find.text('Total'), findsOneWidget);
  });
}
