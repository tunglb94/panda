import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_bottom_sheet.dart';

import '../../domain/models/mock_fare_calculator.dart';
import '../../domain/models/promotion_info.dart';
import '../../domain/models/surge_info.dart';
import '../../domain/models/voucher.dart';

/// Full itemised price breakdown bottom sheet. Every row here maps to a
/// real field on [MockFareBreakdown] (which mirrors `backend/services/
/// pricing`'s formula 1:1) — a component with no backing data source (or
/// worth exactly 0) is omitted entirely rather than shown as a permanent
/// placeholder row, per the production-polish "no placeholder rows" rule.
abstract final class PriceBreakdownSheet {
  static Future<void> show(
    BuildContext context, {
    required MockFareBreakdown fare,
    Voucher? voucher,
    PromotionInfo? promotion,
    SurgeInfo? surge,
  }) {
    return AppBottomSheet.show<void>(
      context,
      title: 'Chi tiết giá',
      isScrollControlled: true,
      builder: (_) => _PriceBreakdownBody(
        fare: fare,
        voucher: voucher,
        promotion: promotion,
        surge: surge,
      ),
    );
  }
}

class _PriceBreakdownBody extends StatelessWidget {
  const _PriceBreakdownBody({
    required this.fare,
    this.voucher,
    this.promotion,
    this.surge,
  });

  final MockFareBreakdown fare;
  final Voucher? voucher;
  final PromotionInfo? promotion;
  final SurgeInfo? surge;

  @override
  Widget build(BuildContext context) {
    // A component that is exactly 0 (e.g. no booking fee configured for
    // this vehicle) or has no backing data source at all (airport/toll/
    // pickup-distance surcharges — no field exists on [MockFareBreakdown])
    // is omitted entirely, never shown as "0 đ" or a permanent placeholder
    // row — no line should ever suggest a charge that isn't real.
    final baseRows = <Widget>[
      if (fare.baseFareCents != 0) _Row(label: 'Giá mở cửa', value: fare.format(fare.baseFareCents)),
      if (fare.distanceFareCents != 0) _Row(label: 'Quãng đường', value: fare.format(fare.distanceFareCents)),
      if (fare.timeFareCents != 0) _Row(label: 'Thời gian', value: fare.format(fare.timeFareCents)),
      if (fare.bookingFeeCents != 0) _Row(label: 'Phí đặt xe', value: fare.format(fare.bookingFeeCents)),
    ];

    final adjustmentRows = <Widget>[
      if (surge != null) _Row(label: 'Giá surge', value: surge!.label, valueColor: AppColors.warning),
      if (voucher != null && fare.discountCents != 0)
        _Row(
          label: 'Voucher',
          value: '-${fare.format(fare.discountCents)} (${voucher!.title})',
          valueColor: AppColors.primary,
        ),
      if (promotion != null)
        _Row(label: 'Khuyến mãi', value: promotion!.title, valueColor: AppColors.primary),
    ];

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          ...baseRows,
          if (baseRows.isNotEmpty && adjustmentRows.isNotEmpty) const _Divider(),
          ...adjustmentRows,
          const _Divider(),
          _Row(
            label: 'Tổng cộng',
            value: fare.format(fare.totalCents),
            isTotal: true,
          ),
        ],
      ),
    );
  }
}

class _Row extends StatelessWidget {
  const _Row({required this.label, required this.value, this.valueColor, this.isTotal = false});

  final String label;
  final String? value;
  final Color? valueColor;
  final bool isTotal;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final labelStyle = isTotal
        ? theme.textTheme.titleSmall
        : theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary);
    final resolvedValue = value ?? 'Chưa áp dụng';
    final valueStyle = isTotal
        ? theme.textTheme.titleLarge?.copyWith(color: AppColors.primary)
        : theme.textTheme.bodyMedium?.copyWith(
            color: value == null ? AppColors.textTertiary : (valueColor ?? AppColors.textPrimary),
            fontWeight: value == null ? FontWeight.w400 : FontWeight.w600,
          );

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs + 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: labelStyle),
          Text(resolvedValue, style: valueStyle),
        ],
      ),
    );
  }
}

class _Divider extends StatelessWidget {
  const _Divider();

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.symmetric(vertical: AppSpacing.xs),
      child: Divider(height: 1, color: AppColors.divider),
    );
  }
}
