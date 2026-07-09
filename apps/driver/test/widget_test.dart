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
import 'package:driver/features/trips/domain/models/dispatch_accept_result.dart';
import 'package:driver/features/trips/domain/models/mock_trip_offer_catalog.dart';
import 'package:driver/features/trips/domain/models/trip_offer_state.dart';
import 'package:driver/features/trips/domain/models/route_progress_model.dart';
import 'package:driver/features/trips/presentation/pages/arrival_preview_page.dart';
import 'package:driver/features/trips/presentation/pages/dispatch_session_preview_page.dart';
import 'package:driver/features/trips/presentation/pages/navigation_preview_page.dart';
import 'package:driver/features/trips/presentation/pages/trip_offer_preview_menu_page.dart';
import 'package:driver/features/trips/presentation/pages/trip_offer_state_preview_page.dart';
import 'package:driver/features/trips/presentation/pages/trips_page.dart';
import 'package:driver/features/trips/presentation/widgets/driver_arrival_card.dart';
import 'package:driver/features/trips/presentation/widgets/driver_navigation_card.dart';
import 'package:driver/features/trips/presentation/widgets/driver_status_banner.dart';
import 'package:driver/features/trips/presentation/widgets/passenger_action_panel.dart';
import 'package:driver/features/trips/presentation/widgets/trip_assigned_card.dart';
import 'package:driver/features/trips/presentation/widgets/waiting_fee_card.dart';
import 'package:driver/features/trips/presentation/widgets/waiting_timer.dart';

void main() {
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

  testWidgets('TripsPage shows loading then a trip offer with countdown',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));

    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    await tester.pump(const Duration(milliseconds: 800));

    expect(find.text('Alex Rider'), findsOneWidget);
    expect(find.text('District 1 Market'), findsOneWidget);
    expect(find.text('Tan Son Nhat Airport'), findsOneWidget);
    expect(find.text('1.5x Surge'), findsOneWidget);
    expect(find.text('15'), findsOneWidget);
    expect(find.text('Accept'), findsOneWidget);
    expect(find.text('Reject'), findsOneWidget);
  });

  testWidgets('TripsPage Reject transitions to Trip Rejected',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await tester.pump(const Duration(milliseconds: 800));

    await tester.tap(find.text('Reject'));
    await tester.pumpAndSettle();

    expect(find.text('Trip Rejected'), findsOneWidget);
    expect(find.text('Accept'), findsNothing);
  });

  testWidgets('Countdown reaching zero automatically shows Offer Expired',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await tester.pump(const Duration(milliseconds: 800));

    expect(find.text('Accept'), findsOneWidget);

    await tester.pump(const Duration(seconds: 16));
    // Let the AnimatedSwitcher's crossfade finish before asserting — right
    // after the countdown fires, the outgoing "New Offer" content and the
    // incoming "Offer Expired" banner briefly coexist mid-transition.
    await tester.pumpAndSettle();

    expect(find.text('Offer Expired'), findsOneWidget);
    expect(find.text('Accept'), findsNothing);
  });

  testWidgets('TripsPage shows empty state via dev preview menu',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.tune));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Empty (dev)'));
    await tester.pumpAndSettle();

    expect(find.text('No incoming trip offers right now'), findsOneWidget);
  });

  testWidgets('TripsPage shows error state via dev preview menu',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.tune));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Error (dev)'));
    await tester.pumpAndSettle();

    expect(find.text("Couldn't load trip offers"), findsOneWidget);
  });

  testWidgets(
      'TripOfferPreviewMenuPage lists the offer states and the Dispatch Session Preview entry',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripOfferPreviewMenuPage()));

    expect(find.text('New Offer'), findsOneWidget);
    expect(find.text('Rejected'), findsOneWidget);
    expect(find.text('Expired'), findsOneWidget);
    expect(find.text('Dispatch Session Preview'), findsOneWidget);
  });

  testWidgets('DispatchSessionPreviewPage lists Accepting/Assigned/Failed/Timeout',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: DispatchSessionPreviewPage()));

    expect(find.text('Accepting'), findsOneWidget);
    expect(find.text('Assigned'), findsOneWidget);
    expect(find.text('Failed'), findsOneWidget);
    expect(find.text('Timeout'), findsOneWidget);
  });

  testWidgets('TripOfferStatePreviewPage renders the Assigned state',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: TripOfferStatePreviewPage(state: TripOfferState.assigned),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Trip Assigned'), findsOneWidget);
    expect(find.text('Start Navigation'), findsOneWidget);
    expect(find.text('Accept'), findsNothing);
  });

  // ─── Phase D-04: Accept Flow & Dispatch Session ──────────────────────────

  testWidgets('TripsPage Accept walks through Accepting to Assigned',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await tester.pump(const Duration(milliseconds: 800));

    await tester.tap(find.text('Accept'));
    await tester.pump();
    // Let the AnimatedSwitcher's ~350ms crossfade finish — right after the
    // tap, the outgoing "New Offer" content (with Reject) and the incoming
    // "Accepting…" content briefly coexist mid-transition.
    await tester.pump(const Duration(milliseconds: 400));

    expect(find.text('Accepting…'), findsOneWidget);
    expect(find.text('Reject'), findsNothing);
    expect(find.text('Accept'), findsNothing);

    await tester.pump(const Duration(milliseconds: 1300));
    await tester.pumpAndSettle();

    expect(find.text('Trip Assigned'), findsOneWidget);
    expect(find.text('Start Navigation'), findsOneWidget);
  });

  testWidgets('The Accept loading button is disabled while accepting',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await tester.pump(const Duration(milliseconds: 800));

    await tester.tap(find.text('Accept'));
    await tester.pump(const Duration(milliseconds: 400));

    final button =
        tester.widget<FilledButton>(find.widgetWithText(FilledButton, 'Accepting…'));
    expect(button.onPressed, isNull);

    // Drain the pending 1.2s accept-delay timer before the test ends —
    // otherwise flutter_test fails this test for leaving a pending Timer,
    // regardless of the assertion above already having passed.
    await tester.pump(const Duration(milliseconds: 1300));
    await tester.pumpAndSettle();
  });

  testWidgets(
      'Countdown is cancelled by Accept and cannot expire the offer afterwards',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await tester.pump(const Duration(milliseconds: 800));

    await tester.tap(find.text('Accept'));
    await tester.pump();
    // Advance well past both the original 15s countdown and the 1.2s accept
    // delay in one jump — if the countdown weren't properly cancelled/guarded,
    // this would race and could flip the state to Expired instead.
    await tester.pump(const Duration(seconds: 20));
    await tester.pumpAndSettle();

    expect(find.text('Trip Assigned'), findsOneWidget);
    expect(find.text('Offer Expired'), findsNothing);
  });

  testWidgets('TripsPage Accept with dev outcome Failed shows the retry banner',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await tester.pump(const Duration(milliseconds: 800));

    // Invoke the dev "Accept outcome" menu's callback directly rather than
    // opening the popup and tapping its item: the offer's 15s countdown is a
    // continuously-ticking `AnimationController` running the whole time this
    // page is on `newOffer`, which makes tap-timing against a popup's own
    // open/close transition unreliable (there is always something animating,
    // so `pumpAndSettle()` cannot be used, and a fixed-duration `pump()` is
    // one animation frame away from a flaky hit-test miss). The menu's own
    // rendering is already covered by the "Preview state (dev)" menu tests.
    _selectAcceptOutcome(tester, DispatchAcceptStatus.failed);

    await tester.tap(find.text('Accept'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 1300));
    await tester.pumpAndSettle();

    expect(find.text('Unable to accept trip.'), findsOneWidget);
    expect(find.text('Try again.'), findsOneWidget);
    expect(find.text('Retry'), findsOneWidget);
  });

  testWidgets('TripsPage Accept with dev outcome Timeout shows the retry banner',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await tester.pump(const Duration(milliseconds: 800));

    _selectAcceptOutcome(tester, DispatchAcceptStatus.timeout);

    await tester.tap(find.text('Accept'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 1300));
    await tester.pumpAndSettle();

    expect(find.text('Dispatch timeout.'), findsOneWidget);
    expect(find.text('Please retry.'), findsOneWidget);
    expect(find.text('Retry'), findsOneWidget);
  });

  testWidgets('Retry after a failed accept returns to a fresh New Offer',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await tester.pump(const Duration(milliseconds: 800));

    _selectAcceptOutcome(tester, DispatchAcceptStatus.failed);

    await tester.tap(find.text('Accept'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 1300));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Retry'));
    // Retry lands back on `newOffer`, which starts a fresh 15s countdown
    // immediately — settle only the ~350ms crossfade, not the whole
    // countdown (see the note above about why `pumpAndSettle()` is unsafe
    // here).
    await tester.pump(const Duration(milliseconds: 400));

    expect(find.text('Accept'), findsOneWidget);
    expect(find.text('Reject'), findsOneWidget);
    expect(find.text('15'), findsOneWidget);
  });

  testWidgets('Assigned screen shows the correct trip data',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await tester.pump(const Duration(milliseconds: 800));

    await tester.tap(find.text('Accept'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 1300));
    await tester.pumpAndSettle();

    expect(find.text('District 1 Market'), findsOneWidget);
    expect(find.text('Tan Son Nhat Airport'), findsOneWidget);
    expect(find.text('\$9.50'), findsOneWidget);
  });

  testWidgets('onNavigate fires exactly once per tap on Start Navigation',
      (WidgetTester tester) async {
    var callCount = 0;
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: TripAssignedCard(
            offer: MockTripOfferCatalog.sample,
            onNavigate: () => callCount++,
          ),
        ),
      ),
    );

    await tester.tap(find.text('Start Navigation'));
    await tester.pump();

    expect(callCount, 1);
  });

  // ─── Phase D-05: Driver Assigned & Navigation Ready ──────────────────────

  testWidgets('Start Navigation transitions Assigned to Navigating to Pickup',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await _reachAssigned(tester);

    expect(find.byType(TripAssignedCard), findsOneWidget);
    expect(find.byType(DriverNavigationCard), findsNothing);

    await tester.tap(find.text('Start Navigation'));
    await _reachNavigatingContent(tester);

    expect(find.text('Driving to Pickup'), findsOneWidget);
    expect(find.byType(TripAssignedCard), findsNothing);
    expect(find.byType(DriverNavigationCard), findsOneWidget);

    // The route-progress ticker (100 -> 0, every 2s) is still running —
    // unmount and drain its one pending tick before the test ends, same
    // "pending Timer" discipline as the accept-delay tests above.
    await _disposeAndDrainPendingTimer(tester);
  });

  testWidgets('Navigation screen renders initial distance, ETA and traffic',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await _reachAssigned(tester);

    await tester.tap(find.text('Start Navigation'));
    await _reachNavigatingContent(tester);

    expect(find.text('1.8 km'), findsOneWidget);
    expect(find.text('6 min'), findsOneWidget);
    expect(find.text('100% remaining'), findsOneWidget);
    expect(find.text('Normal traffic'), findsOneWidget);

    await _disposeAndDrainPendingTimer(tester);
  });

  testWidgets('Route progress updates over time (100% -> 80%)',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await _reachAssigned(tester);

    await tester.tap(find.text('Start Navigation'));
    await _reachNavigatingContent(tester);

    expect(find.text('100% remaining'), findsOneWidget);

    await tester.pump(const Duration(milliseconds: 2100));

    expect(find.text('80% remaining'), findsOneWidget);
    expect(find.text('1.4 km'), findsOneWidget);
    expect(find.text('5 min'), findsOneWidget);

    await _disposeAndDrainPendingTimer(tester);
  });

  testWidgets('Traffic badge reflects the dev-selected traffic level',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await _reachAssigned(tester);

    _selectTraffic(tester, TrafficLevel.heavy);

    await tester.tap(find.text('Start Navigation'));
    await _reachNavigatingContent(tester);

    expect(find.text('Heavy traffic'), findsOneWidget);
    expect(find.text('10 min'), findsOneWidget);
    expect(find.text('1.8 km'), findsOneWidget);

    await _disposeAndDrainPendingTimer(tester);
  });

  testWidgets('Cancel Trip is a plain callback with no popup',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await _reachAssigned(tester);

    await tester.tap(find.text('Start Navigation'));
    await _reachNavigatingContent(tester);

    await tester.tap(find.text('Cancel Trip'));
    await tester.pump();

    expect(find.text('Cancel Trip is a placeholder — not yet implemented.'), findsOneWidget);
    expect(find.byType(AlertDialog), findsNothing);
    expect(find.byType(Dialog), findsNothing);

    await _disposeAndDrainPendingTimer(tester);
  });

  testWidgets('Contact Rider is a plain callback with no popup',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await _reachAssigned(tester);

    await tester.tap(find.text('Start Navigation'));
    await _reachNavigatingContent(tester);

    await tester.tap(find.text('Contact Rider'));
    await tester.pump();

    expect(find.text('Contact Rider is a placeholder — not yet implemented.'), findsOneWidget);
    expect(find.byType(AlertDialog), findsNothing);
    expect(find.byType(Dialog), findsNothing);

    await _disposeAndDrainPendingTimer(tester);
  });

  testWidgets(
      'TripOfferPreviewMenuPage lists the Navigation Preview entry',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripOfferPreviewMenuPage()));

    expect(find.text('Navigation Preview'), findsOneWidget);
  });

  testWidgets('NavigationPreviewPage steps through Assigned -> 100% -> 20% without a repository',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: NavigationPreviewPage()));
    await tester.pump();

    expect(find.text('Trip Assigned'), findsOneWidget);
    expect(find.text('Start Navigation'), findsOneWidget);

    await tester.tap(find.widgetWithText(ChoiceChip, '100%'));
    await tester.pump(const Duration(milliseconds: 400));

    expect(find.text('Driving to Pickup'), findsOneWidget);
    expect(find.text('100% remaining'), findsOneWidget);

    await tester.tap(find.widgetWithText(ChoiceChip, '20%'));
    await tester.pump(const Duration(milliseconds: 400));

    expect(find.text('20% remaining'), findsOneWidget);

    // No repository/Future involved in the preview — nothing pending to
    // drain, unlike the live-flow tests above.
  });

  // ─── Phase D-06: Arrived at Pickup ───────────────────────────────────────

  testWidgets(
      'Route progress completing transitions Navigating to Arrived at Pickup',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripsPage()));
    await _reachAssigned(tester);

    await tester.tap(find.text('Start Navigation'));
    await _reachNavigatingContent(tester);

    expect(find.byType(DriverNavigationCard), findsOneWidget);
    expect(find.byType(DriverArrivalCard), findsNothing);

    await _reachArrivedContent(tester);

    expect(find.text('Arrived at Pickup'), findsOneWidget);
    expect(find.byType(DriverNavigationCard), findsNothing);
    expect(find.byType(DriverArrivalCard), findsOneWidget);

    await _disposeAndDrainPendingTimer(tester);
  });

  testWidgets('WaitingTimer counts up in mm:ss and fires onMinutePassed',
      (WidgetTester tester) async {
    final minutes = <int>[];
    await tester.pumpWidget(
      MaterialApp(home: Scaffold(body: WaitingTimer(onMinutePassed: minutes.add))),
    );

    expect(find.text('00:00'), findsOneWidget);

    await tester.pump(const Duration(seconds: 1));
    expect(find.text('00:01'), findsOneWidget);

    await tester.pump(const Duration(seconds: 59));
    expect(find.text('01:00'), findsOneWidget);
    expect(minutes, [1]);

    await _disposeAndDrainPendingTimer(tester);
  });

  testWidgets('WaitingTimer does not tick after being disposed',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: Scaffold(body: WaitingTimer())));
    await tester.pump(const Duration(seconds: 3));

    expect(find.text('00:03'), findsOneWidget);

    // Unmount — dispose() flips the `_stopped` guard. If it didn't, the
    // pump below would crash with "setState() called after dispose()"
    // instead of just quietly resolving the one already-scheduled tick.
    await tester.pumpWidget(const SizedBox());
    await tester.pump(const Duration(seconds: 5));
  });

  testWidgets('WaitingFeeCard is free for 5 minutes then charges per minute',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: WaitingFeeCard(elapsedMinutes: 5))),
    );
    expect(find.text('Free'), findsOneWidget);

    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: WaitingFeeCard(elapsedMinutes: 6))),
    );
    expect(find.text('2.000đ'), findsOneWidget);

    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: WaitingFeeCard(elapsedMinutes: 8))),
    );
    expect(find.text('6.000đ'), findsOneWidget);
  });

  testWidgets('PassengerActionPanel buttons fire their respective callbacks',
      (WidgetTester tester) async {
    var onBoardCount = 0;
    var contactCount = 0;
    var cancelCount = 0;
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PassengerActionPanel(
            onPassengerOnBoard: () => onBoardCount++,
            onContactRider: () => contactCount++,
            onCancelTrip: () => cancelCount++,
          ),
        ),
      ),
    );

    await tester.tap(find.text('Passenger On Board'));
    await tester.tap(find.text('Contact Rider'));
    await tester.tap(find.text('Cancel Trip'));
    await tester.pump();

    expect(onBoardCount, 1);
    expect(contactCount, 1);
    expect(cancelCount, 1);
  });

  testWidgets(
      'DriverArrivalCard shows the Arrived at Pickup status banner via DriverStatusBanner',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: DriverArrivalCard(
            offer: MockTripOfferCatalog.sample,
            onPassengerOnBoard: () {},
            onContactRider: () {},
            onCancelTrip: () {},
          ),
        ),
      ),
    );

    expect(find.text('Arrived at Pickup'), findsOneWidget);
    expect(find.byType(DriverStatusBanner), findsOneWidget);
    expect(find.text('District 1 Market'), findsOneWidget);
    expect(find.text('Alex Rider'), findsOneWidget);

    await _disposeAndDrainPendingTimer(tester);
  });

  testWidgets('TripOfferPreviewMenuPage lists the Arrival Preview entry',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: TripOfferPreviewMenuPage()));

    expect(find.text('Arrival Preview'), findsOneWidget);
  });

  testWidgets(
      'ArrivalPreviewPage steps through Arrived/Waiting durations without a repository',
      (WidgetTester tester) async {
    await tester.pumpWidget(const MaterialApp(home: ArrivalPreviewPage()));
    await tester.pump();

    expect(find.text('00:00'), findsOneWidget);

    await tester.tap(find.widgetWithText(ChoiceChip, 'Waiting 03:00'));
    await tester.pump(const Duration(milliseconds: 400));

    expect(find.text('03:00'), findsOneWidget);

    await tester.tap(find.widgetWithText(ChoiceChip, 'Waiting 08:00'));
    await tester.pump(const Duration(milliseconds: 400));

    expect(find.text('08:00'), findsOneWidget);

    await _disposeAndDrainPendingTimer(tester);
  });
}

/// Selects [outcome] on `TripsPage`'s "Accept outcome (dev)" menu by calling
/// its `onSelected` callback directly, bypassing the popup menu's own open/
/// tap/close UI interaction. See the comment at the first call site for why:
/// the offer's 15s countdown keeps an `AnimationController` running the
/// entire time the page shows `newOffer`, which makes `pumpAndSettle()`
/// unsafe and fixed-duration `pump()`s one frame away from a flaky
/// popup-animation hit-test miss.
void _selectAcceptOutcome(WidgetTester tester, DispatchAcceptStatus outcome) {
  final button = tester.widget<PopupMenuButton<DispatchAcceptStatus>>(
    find.byType(PopupMenuButton<DispatchAcceptStatus>),
  );
  button.onSelected!(outcome);
}

/// Selects [traffic] on `TripsPage`'s "Traffic (dev)" menu the same
/// direct-callback way as `_selectAcceptOutcome` — see that function's
/// comment for why (the offer's 15s countdown keeps an `AnimationController`
/// running, which makes tap-through-popup timing unreliable).
void _selectTraffic(WidgetTester tester, TrafficLevel traffic) {
  final button = tester.widget<PopupMenuButton<TrafficLevel>>(
    find.byType(PopupMenuButton<TrafficLevel>),
  );
  button.onSelected!(traffic);
}

/// Drives `TripsPage` from a freshly pumped New Offer through Accept to the
/// Assigned screen. Shared setup for the Phase D-05 navigation tests below.
Future<void> _reachAssigned(WidgetTester tester) async {
  await tester.pump(const Duration(milliseconds: 800));
  await tester.tap(find.text('Accept'));
  await tester.pump();
  await tester.pump(const Duration(milliseconds: 1300));
  await tester.pumpAndSettle();
}

/// After tapping "Start Navigation", advances past the `AnimatedSwitcher`
/// crossfade (~350ms) and the mock `fetchRouteProgress` delay (600ms) so the
/// initial `DriverNavigationCard` (100%) is showing. Deliberately does NOT
/// use `pumpAndSettle()` — see `_RouteProgressTicker`'s doc comment: once
/// this content is showing, a 2s-spaced tick chain is already scheduled, and
/// `pumpAndSettle()`'s short repeated pumps don't reliably advance a bare
/// `Future.delayed` the way one big `pump(duration)` does.
Future<void> _reachNavigatingContent(WidgetTester tester) async {
  await tester.pump(const Duration(milliseconds: 400));
  await tester.pump(const Duration(milliseconds: 700));
}

/// From the initial (100%) `DriverNavigationCard`, advances past all 5
/// route-progress ticks (100 -> 80 -> 60 -> 40 -> 20 -> 0, 2s apart — one
/// big `pump` fires the whole chain, same reasoning as `_reachNavigatingContent`)
/// so `onArrived` fires and the state flips to `arrivedAtPickup`, then lets
/// the `AnimatedSwitcher` crossfade into `DriverArrivalCard` finish.
Future<void> _reachArrivedContent(WidgetTester tester) async {
  await tester.pump(const Duration(seconds: 11));
  await tester.pump(const Duration(milliseconds: 400));
}

/// Both `_RouteProgressTicker` and `WaitingTimer` (Phase D-06) reschedule
/// themselves via a bare `Future.delayed` chain, and a test that isn't
/// specifically driving one of them all the way to its end (e.g. the D-05
/// tests above, which only care about the *initial* Navigating content, not
/// about it eventually reaching Arrived) can't just `pump()` past it — that
/// would either take the test somewhere it doesn't want to go (`arrivedAtPickup`,
/// cascading into a perpetual `WaitingTimer` of its own) or, for `WaitingTimer`,
/// never finish at all (it has no floor). The reliable fix either way: unmount
/// the tree (its `dispose()` flips the `_stopped` guard so no further
/// re-scheduling happens), then pump past the single tick that was already
/// scheduled before disposal so it fires harmlessly instead of leaving a
/// "pending Timer" failure at test end.
Future<void> _disposeAndDrainPendingTimer(WidgetTester tester) async {
  await tester.pumpWidget(const SizedBox());
  await tester.pump(const Duration(seconds: 3));
}
