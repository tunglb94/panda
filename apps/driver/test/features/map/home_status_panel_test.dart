// Widget tests for Phần 7 (Driver KYC spec) — the Home Online switch must
// be disabled with a "Cần hoàn thành xác minh." tooltip whenever the
// driver's KYC + Vehicle Verification aren't both Approved yet.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:driver/features/map/presentation/widgets/home_status_panel.dart';

void main() {
  testWidgets('Switch is enabled when canGoOnline is true (offline phase)', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: HomeStatusPanel(
            phase: HomePhase.offline,
            isBusy: false,
            canGoOnline: true,
            onToggleOnline: () {},
            onViewTrip: () {},
          ),
        ),
      ),
    );
    await tester.pump();

    final switchWidget = tester.widget<Switch>(find.byType(Switch));
    expect(switchWidget.onChanged, isNotNull);

    final tooltip = tester.widget<Tooltip>(find.byType(Tooltip));
    expect(tooltip.message, isNot(contains('xác minh')));
  });

  testWidgets('Switch is disabled with verification tooltip when canGoOnline is false', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: HomeStatusPanel(
            phase: HomePhase.offline,
            isBusy: false,
            canGoOnline: false,
            onToggleOnline: () {},
            onViewTrip: () {},
          ),
        ),
      ),
    );
    await tester.pump();

    final switchWidget = tester.widget<Switch>(find.byType(Switch));
    expect(switchWidget.onChanged, isNull);

    final tooltip = tester.widget<Tooltip>(find.byType(Tooltip));
    expect(tooltip.message, 'Cần hoàn thành xác minh.');

    final semanticsMatch = find.byWidgetPredicate(
      (w) => w is Semantics && (w.properties.label?.contains('xác minh') ?? false),
    );
    expect(semanticsMatch, findsOneWidget);
  });

  testWidgets('Going offline is always allowed even when canGoOnline is false', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: HomeStatusPanel(
            phase: HomePhase.online,
            isBusy: false,
            canGoOnline: false,
            onToggleOnline: () {},
            onViewTrip: () {},
          ),
        ),
      ),
    );
    await tester.pump();

    final switchWidget = tester.widget<Switch>(find.byType(Switch));
    expect(switchWidget.onChanged, isNotNull);
    expect(switchWidget.value, isTrue);
  });
}
