// Widget tests for Phần 8 (Driver KYC spec) — the Rider Contact Card must
// show the real KYC verified badge (from `ContactInfo.isVerified`, backed
// by DriverVerification+VehicleVerification both Approved), join date, and
// trip count — and must never show/derive from the legacy
// `DriverProfile.isVerified` (a different, older field) or any raw document.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:rider/features/contact/domain/models/contact_info.dart';
import 'package:rider/features/trip/domain/models/driver_profile.dart';
import 'package:rider/features/trip/presentation/widgets/driver_info_card.dart';

const _driver = DriverProfile(
  vehicleBrand: 'Honda',
  vehicleModel: 'Wave',
  vehicleColor: 'Đỏ',
  plateNumber: '59-X1 123.45',
  verificationStatus: '', // legacy field intentionally left unset
);

void main() {
  testWidgets('shows verified badge, trip count, and join year from ContactInfo', (tester) async {
    final contact = ContactInfo(
      name: 'Trần Văn B',
      maskedPhone: '090****123',
      rating: 4.8,
      ratingCount: 120,
      isVerified: true,
      joinedAt: DateTime.utc(2024, 3, 15),
      tripCount: 342,
    );

    await tester.pumpWidget(
      MaterialApp(home: Scaffold(body: DriverInfoCard(driver: _driver, contact: contact))),
    );

    expect(find.byIcon(Icons.verified), findsOneWidget);
    expect(find.text('342 chuyến'), findsOneWidget);
    expect(find.text('Tham gia từ 2024'), findsOneWidget);
    expect(find.text('4.8'), findsOneWidget);
  });

  testWidgets('shows no verified badge when contact is null (legacy field ignored)', (tester) async {
    await tester.pumpWidget(
      MaterialApp(home: Scaffold(body: DriverInfoCard(driver: _driver))),
    );

    expect(find.byIcon(Icons.verified), findsNothing);
    expect(find.textContaining('chuyến'), findsNothing);
    expect(find.textContaining('Tham gia từ'), findsNothing);
  });
}
