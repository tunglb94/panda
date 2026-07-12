import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';

import '../../domain/models/promotion_info.dart';

/// Banner shown when a promotion applies to this trip — an emoji, a title,
/// and the specific reason it applies (never a generic "you have an offer").
///
/// Renders nothing when [promotion] is null. No promotion source is wired
/// up anywhere in the app today (see `PromotionInfo`'s doc comment), so in
/// practice this widget is currently always given `null` — it is fully
/// built and ready for the day a real promotion is available to pass in.
class PromotionBanner extends StatelessWidget {
  const PromotionBanner({super.key, required this.promotion});

  final PromotionInfo? promotion;

  @override
  Widget build(BuildContext context) {
    final promo = promotion;
    if (promo == null) return const SizedBox.shrink();

    return AnimatedOpacity(
      opacity: 1,
      duration: const Duration(milliseconds: 260),
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(AppSpacing.md),
        decoration: BoxDecoration(
          color: AppColors.primaryLight,
          borderRadius: AppRadius.mdAll,
          border: Border.all(color: AppColors.primary.withValues(alpha: 0.24)),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(promo.kind.emoji, style: const TextStyle(fontSize: AppIconSize.xl)),
            const SizedBox(width: AppSpacing.sm),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    promo.title,
                    style: Theme.of(context).textTheme.titleSmall?.copyWith(color: AppColors.primaryDark),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    promo.reason,
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
