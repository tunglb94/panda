// Tests for Section 3 (Promotion Explanation) of the Payment/Fare
// production pass: an applied voucher must show its code, amount, and
// reason; an ineligible voucher must show why, in plain text — never a
// generic "not eligible" with no explanation.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:rider/features/booking/domain/models/voucher.dart';
import 'package:rider/features/booking/presentation/widgets/voucher_card.dart';

void main() {
  testWidgets('Applied voucher shows code, discount, and reason ("Đã áp dụng")',
      (tester) async {
    const voucher = Voucher(
      id: 'v1',
      code: 'FIRST50',
      title: 'Ưu đãi chuyến đầu tiên',
      description: 'Giảm 50% tối đa 20.000đ',
      icon: Icons.celebration,
      accentColor: Colors.orange,
      discountLabel: '-20.000đ',
      status: VoucherStatus.applied,
      conditionText: 'Chuyến đầu tiên',
      discountPercent: 50,
    );

    await tester.pumpWidget(const MaterialApp(home: Scaffold(body: VoucherCard(voucher: voucher))));

    expect(find.textContaining('Đã áp dụng'), findsOneWidget);
    expect(find.textContaining('FIRST50'), findsWidgets);
    expect(find.textContaining('Lý do: Chuyến đầu tiên'), findsOneWidget);
  });

  testWidgets('Unavailable voucher shows "Không đủ điều kiện" with the exact reason',
      (tester) async {
    const voucher = Voucher(
      id: 'v2',
      code: 'MIN80',
      title: 'Giảm 30.000đ',
      description: 'Cho đơn từ 80.000đ',
      icon: Icons.local_offer,
      accentColor: Colors.blue,
      discountLabel: '-30.000đ',
      status: VoucherStatus.unavailable,
      conditionText: 'Đơn tối thiểu 80.000đ',
    );

    await tester.pumpWidget(const MaterialApp(home: Scaffold(body: VoucherCard(voucher: voucher))));

    expect(find.textContaining('Không đủ điều kiện'), findsOneWidget);
    expect(find.textContaining('Đơn tối thiểu 80.000đ'), findsOneWidget);
  });

  testWidgets('Unavailable voucher for wrong city/vehicle/exhausted usage all render honestly',
      (tester) async {
    for (final reason in ['Không đúng thành phố', 'Hết lượt sử dụng', 'Sai loại xe']) {
      final voucher = Voucher(
        id: 'v3',
        code: 'X',
        title: 'X',
        description: 'X',
        icon: Icons.local_offer,
        accentColor: Colors.blue,
        discountLabel: '-10%',
        status: VoucherStatus.unavailable,
        conditionText: reason,
      );
      await tester.pumpWidget(MaterialApp(home: Scaffold(body: VoucherCard(voucher: voucher))));
      expect(find.textContaining(reason), findsOneWidget, reason: 'expected reason "$reason" to be shown verbatim');
    }
  });
}
