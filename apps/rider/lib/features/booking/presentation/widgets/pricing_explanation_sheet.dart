import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_bottom_sheet.dart';

import '../../domain/models/mock_fare_calculator.dart';
import '../../domain/models/pricing_explanation.dart';
import '../../domain/models/surge_info.dart';
import '../../domain/models/voucher.dart';

/// "Tại sao giá này?" — PHẦN 2 of the Payment/Fare production-readiness
/// pass. A checklist-style bottom sheet answering, line by line, why the
/// estimate shows the amount it does. See [PricingExplanation.build] for
/// exactly how each line is derived (rule-based, not AI).
abstract final class PricingExplanationSheet {
  static Future<void> show(
    BuildContext context, {
    required MockFareBreakdown fare,
    required double distanceKm,
    required double durationMin,
    Voucher? voucher,
    SurgeInfo? surge,
  }) {
    final lines = PricingExplanation.build(
      fare: fare,
      distanceKm: distanceKm,
      durationMin: durationMin,
      requestTime: DateTime.now(),
      voucher: voucher,
      surge: surge,
    );
    return AppBottomSheet.show<void>(
      context,
      title: 'Tại sao giá này?',
      builder: (_) => _PricingExplanationBody(lines: lines, fare: fare),
    );
  }
}

class _PricingExplanationBody extends StatelessWidget {
  const _PricingExplanationBody({required this.lines, required this.fare});

  final List<PricingExplanationLine> lines;
  final MockFareBreakdown fare;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    // [PricingExplanation.build] always appends exactly these 3 lines last,
    // in this order — pulled out to render as chips (rule/surge/promotion)
    // instead of plain checklist rows; the leading lines (base fare,
    // distance, duration) stay as the checklist. Presentation-only split,
    // the underlying data/order from the builder is unchanged.
    final hasChipLines = lines.length >= 3;
    final checklistLines = hasChipLines ? lines.sublist(0, lines.length - 3) : lines;
    final chipLines = hasChipLines ? lines.sublist(lines.length - 3) : const <PricingExplanationLine>[];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          'Đây là ước tính, tính từ quãng đường, thời gian di chuyển và các phụ phí hiện có.',
          style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
        ),
        const SizedBox(height: AppSpacing.lg),
        for (final line in checklistLines)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs + 2),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(
                  line.isPositive ? Icons.check_circle : Icons.info,
                  size: 18,
                  color: line.isPositive ? AppColors.primary : AppColors.warning,
                ),
                const SizedBox(width: AppSpacing.sm),
                Expanded(
                  child: Text(line.text, style: theme.textTheme.bodyMedium),
                ),
              ],
            ),
          ),
        if (chipLines.isNotEmpty) ...[
          const Padding(
            padding: EdgeInsets.symmetric(vertical: AppSpacing.md),
            child: Divider(height: 1, color: AppColors.divider),
          ),
          Wrap(
            spacing: AppSpacing.sm,
            runSpacing: AppSpacing.sm,
            children: [for (final line in chipLines) _ExplanationChip(line: line)],
          ),
        ],
        const Padding(
          padding: EdgeInsets.symmetric(vertical: AppSpacing.lg),
          child: Divider(height: 1, color: AppColors.divider),
        ),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text('Tổng cộng', style: theme.textTheme.titleSmall),
            Flexible(
              child: Text(
                fare.format(fare.totalCents),
                textAlign: TextAlign.right,
                style: theme.textTheme.titleLarge?.copyWith(color: AppColors.primary),
              ),
            ),
          ],
        ),
      ],
    );
  }
}

/// A rule/surge/promotion fact rendered as a small pill instead of a plain
/// checklist row — purely a visual regrouping of the same
/// [PricingExplanationLine] data, no new information.
class _ExplanationChip extends StatelessWidget {
  const _ExplanationChip({required this.line});

  final PricingExplanationLine line;

  @override
  Widget build(BuildContext context) {
    final color = line.isPositive ? AppColors.primary : AppColors.warning;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            line.isPositive ? Icons.check_circle_outline : Icons.info_outline,
            size: 16,
            color: color,
          ),
          const SizedBox(width: 6),
          Flexible(
            child: Text(
              line.text,
              style: Theme.of(context).textTheme.labelMedium?.copyWith(color: color),
            ),
          ),
        ],
      ),
    );
  }
}
