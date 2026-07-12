import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/animated_counter.dart';

/// "Giá trước → Giá sau ưu đãi → Tiết kiệm", each figure animating in with a
/// count-up. Renders nothing when there is no discount to show (original
/// and final price equal) — only real, currently-computed numbers are ever
/// shown; there is no separate "price history over time" data source, so
/// this widget represents the current estimate's before/after, not a
/// multi-trip timeline.
class PriceHistoryWidget extends StatelessWidget {
  const PriceHistoryWidget({
    super.key,
    required this.originalCents,
    required this.finalCents,
    required this.format,
  });

  final int originalCents;
  final int finalCents;
  final String Function(int) format;

  @override
  Widget build(BuildContext context) {
    if (originalCents <= finalCents) return const SizedBox.shrink();
    final savedCents = originalCents - finalCents;

    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 260),
      child: Column(
        key: const ValueKey('price-history'),
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _row(
            context,
            label: 'Giá trước',
            child: Text(
              format(originalCents),
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: AppColors.textTertiary,
                    decoration: TextDecoration.lineThrough,
                  ),
            ),
          ),
          _arrow(),
          _row(
            context,
            label: 'Giá sau ưu đãi',
            child: AnimatedCounter(
              value: finalCents,
              format: format,
              style: Theme.of(context)
                  .textTheme
                  .titleSmall
                  ?.copyWith(color: AppColors.primary),
            ),
          ),
          _arrow(),
          _row(
            context,
            label: 'Bạn tiết kiệm',
            child: AnimatedCounter(
              value: savedCents,
              format: (v) => '-${format(v)}',
              style: Theme.of(context).textTheme.titleSmall?.copyWith(color: AppColors.primary),
            ),
          ),
        ],
      ),
    );
  }

  Widget _row(BuildContext context, {required String label, required Widget child}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: Theme.of(context).textTheme.bodySmall),
          child,
        ],
      ),
    );
  }

  Widget _arrow() {
    return const Icon(Icons.arrow_downward, size: AppIconSize.sm, color: AppColors.textTertiary);
  }
}
