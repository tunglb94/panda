// Tests for PricingExplanation ("Tại sao giá này?" — Section 2/11 of the
// Payment/Fare production pass): rule-based, deterministic, not AI —
// covers the Voucher / No Voucher / Promotion-adjacent checklist lines.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:rider/features/booking/domain/models/fare_estimate.dart';
import 'package:rider/features/booking/domain/models/pricing_explanation.dart';
import 'package:rider/features/booking/domain/models/surge_info.dart';
import 'package:rider/features/booking/domain/models/voucher.dart';

const _fare = FareEstimate(
  serviceType: 'car',
  vehicleType: 'car',
  distanceKm: 8.3,
  durationMinutes: 18,
  baseFare: 10000,
  distanceFare: 33200,
  timeFare: 7200,
  bookingFee: 2000,
  rideFare: 50400,
  total: 52400,
  currencyCode: 'VND',
  isFinal: false,
);

void main() {
  group('PricingExplanation.build', () {
    test('No Voucher: shows "Không áp dụng voucher" and no discount line', () {
      final lines = PricingExplanation.build(
        fare: _fare,
        distanceKm: 8.3,
        durationMin: 18,
        requestTime: DateTime(2026, 1, 5, 22), // Monday 22:00 — off-peak
        voucher: null,
        surge: null,
      );

      expect(lines.map((l) => l.text), contains('Không áp dụng voucher'));
      expect(lines.any((l) => l.text.startsWith('Voucher') || l.text.startsWith('Áp dụng mã')), isFalse);
    });

    test('Voucher: shows the voucher code was applied (no discount amount — backend has none)', () {
      const voucher = Voucher(
        id: 'v1',
        code: 'FIRST50',
        title: 'Chuyến đầu tiên',
        description: 'Giảm 50% cho chuyến đầu tiên',
        icon: Icons.celebration,
        accentColor: Colors.orange,
        discountLabel: '-50%',
        status: VoucherStatus.applied,
      );
      final lines = PricingExplanation.build(
        fare: _fare,
        distanceKm: 8.3,
        durationMin: 18,
        requestTime: DateTime(2026, 1, 5, 22),
        voucher: voucher,
        surge: null,
      );

      final voucherLine = lines.firstWhere((l) => l.text.startsWith('Áp dụng mã'));
      expect(voucherLine.text, contains('FIRST50'));
    });

    test('distance/duration lines reflect the real trip geometry', () {
      final lines = PricingExplanation.build(
        fare: _fare,
        distanceKm: 8.3,
        durationMin: 18,
        requestTime: DateTime(2026, 1, 5, 22),
      );
      expect(lines.map((l) => l.text), contains('8.3 km'));
      expect(lines.map((l) => l.text), contains('18 phút'));
    });

    test('peak hour window (BRB §2.2.12): weekday 08:00 is flagged as peak', () {
      final lines = PricingExplanation.build(
        fare: _fare,
        distanceKm: 8.3,
        durationMin: 18,
        requestTime: DateTime(2026, 1, 5, 8, 0), // Monday 08:00
      );
      expect(lines.any((l) => l.text.contains('Trong khung giờ cao điểm')), isTrue);
    });

    test('peak hour window (BRB §2.2.12): weekday 14:00 is not peak', () {
      final lines = PricingExplanation.build(
        fare: _fare,
        distanceKm: 8.3,
        durationMin: 18,
        requestTime: DateTime(2026, 1, 5, 14, 0), // Monday 14:00
      );
      expect(lines.any((l) => l.text == 'Không trong giờ cao điểm'), isTrue);
    });

    test('peak hour window (BRB §2.2.12): Saturday 08:00 is not peak (weekday-only rule)', () {
      final lines = PricingExplanation.build(
        fare: _fare,
        distanceKm: 8.3,
        durationMin: 18,
        requestTime: DateTime(2026, 1, 3, 8, 0), // Saturday 08:00
      );
      expect(lines.any((l) => l.text == 'Không trong giờ cao điểm'), isTrue);
    });

    test('No surge: shows "Không áp dụng Surge"', () {
      final lines = PricingExplanation.build(
        fare: _fare,
        distanceKm: 8.3,
        durationMin: 18,
        requestTime: DateTime(2026, 1, 5, 22),
        surge: null,
      );
      expect(lines.map((l) => l.text), contains('Không áp dụng Surge'));
    });

    test('Surge present: shows the surge label and is flagged non-positive', () {
      const surge = SurgeInfo(label: 'Giá đang thay đổi', explanation: 'Nhu cầu tăng cao tạm thời.');
      final lines = PricingExplanation.build(
        fare: _fare,
        distanceKm: 8.3,
        durationMin: 18,
        requestTime: DateTime(2026, 1, 5, 22),
        surge: surge,
      );
      final surgeLine = lines.firstWhere((l) => l.text.contains('Giá đang thay đổi'));
      expect(surgeLine.isPositive, isFalse);
    });
  });
}
