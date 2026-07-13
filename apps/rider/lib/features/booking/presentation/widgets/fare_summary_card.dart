import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/animated_counter.dart';
import 'package:rider/shared/widgets/app_card.dart';
import 'package:rider/shared/widgets/app_status_chip.dart';

import 'package:rider/shared/utils/currency_format.dart';

import '../../domain/models/fare_estimate.dart';
import '../../domain/models/promotion_info.dart';
import '../../domain/models/surge_info.dart';
import '../../domain/models/voucher.dart';
import 'price_breakdown_sheet.dart';
import 'pricing_explanation_sheet.dart';
import 'promotion_banner.dart';
import 'surge_indicator.dart';

/// Trip price summary card: total price, applied voucher, promotion reason,
/// and a competitor-price badge — each shown only when there is a real
/// value backing it (see each field's doc comment). Tap "Chi tiết" to open
/// the full itemised [PriceBreakdownSheet].
///
/// [breakdown] is the real fare returned by the backend's Pricing service
/// (`PricingRepository.estimateFare`) — Flutter performs no fare math.
/// Always labelled "ước tính" (estimate), never presented as a
/// confirmed/final price (`isFinal` is always false for a pre-booking quote).
class FareSummaryCard extends StatelessWidget {
  const FareSummaryCard({
    super.key,
    required this.breakdown,
    required this.distanceKm,
    required this.durationMin,
    this.voucher,
    this.promotion,
    this.surge,
    this.cheaperThanCompetitorLabel,
  });

  final FareEstimate breakdown;

  /// Real trip geometry, used only to render the "Tại sao giá này?"
  /// explanation; not itself an invented pricing input.
  final double distanceKm;
  final double durationMin;

  final Voucher? voucher;
  final PromotionInfo? promotion;
  final SurgeInfo? surge;

  /// e.g. "Rẻ hơn Grab 12%". Null hides the badge entirely — there is no
  /// competitor-price data source anywhere in the backend today, so this is
  /// always null in practice; the badge is built and ready per the sprint
  /// brief's "nếu backend trả được" condition.
  final String? cheaperThanCompetitorLabel;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.receipt_long_outlined, size: AppIconSize.md, color: AppColors.primary),
              const SizedBox(width: AppSpacing.sm),
              Expanded(
                child: Text('Giá chuyến đi', style: Theme.of(context).textTheme.titleSmall),
              ),
              GestureDetector(
                onTap: () => PriceBreakdownSheet.show(
                  context,
                  fare: breakdown,
                  voucher: voucher,
                  promotion: promotion,
                  surge: surge,
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      'Chi tiết',
                      style: Theme.of(context).textTheme.labelLarge?.copyWith(color: AppColors.primary),
                    ),
                    Icon(Icons.chevron_right, size: AppIconSize.sm, color: AppColors.primary),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.xs),
          Align(
            alignment: Alignment.centerLeft,
            child: GestureDetector(
              onTap: () => PricingExplanationSheet.show(
                context,
                fare: breakdown,
                distanceKm: distanceKm,
                durationMin: durationMin,
                voucher: voucher,
                surge: surge,
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.help_outline, size: AppIconSize.sm, color: AppColors.textSecondary),
                  const SizedBox(width: 4),
                  Text(
                    'Tại sao giá này?',
                    style: Theme.of(context).textTheme.labelMedium?.copyWith(
                          color: AppColors.textSecondary,
                          decoration: TextDecoration.underline,
                        ),
                  ),
                ],
              ),
            ),
          ),
          if (promotion != null) ...[
            const SizedBox(height: AppSpacing.md),
            PromotionBanner(promotion: promotion),
          ],
          const SizedBox(height: AppSpacing.md),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(AppSpacing.md),
            decoration: BoxDecoration(
              color: AppColors.primaryLight,
              borderRadius: AppRadius.mdAll,
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Flexible(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          AnimatedCounter(
                            value: breakdown.total,
                            format: (v) => formatMoney(v, breakdown.currencyCode),
                            style: Theme.of(context).textTheme.headlineSmall?.copyWith(color: AppColors.primary),
                          ),
                          Text('Giá ước tính', style: Theme.of(context).textTheme.labelSmall),
                        ],
                      ),
                    ),
                    if (cheaperThanCompetitorLabel != null)
                      AppStatusChip(
                        label: cheaperThanCompetitorLabel!,
                        color: AppColors.primary,
                        icon: Icons.trending_down,
                      ),
                  ],
                ),
              ],
            ),
          ),
          if (surge != null || voucher != null) ...[
            const SizedBox(height: AppSpacing.sm),
            Wrap(
              spacing: AppSpacing.sm,
              runSpacing: AppSpacing.sm,
              children: [
                if (surge != null) SurgeIndicator(surge: surge),
                if (voucher != null)
                  AppStatusChip(
                    label: '${voucher!.title} · ${voucher!.discountLabel}',
                    color: AppColors.primary,
                    icon: Icons.local_offer,
                  ),
              ],
            ),
          ],
        ],
      ),
    );
  }
}
