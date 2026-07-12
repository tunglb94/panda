import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/animated_counter.dart';
import 'package:rider/shared/widgets/app_card.dart';
import 'package:rider/shared/widgets/app_status_chip.dart';

import '../../domain/models/mock_fare_calculator.dart';
import '../../domain/models/promotion_info.dart';
import '../../domain/models/surge_info.dart';
import '../../domain/models/voucher.dart';
import 'price_breakdown_sheet.dart';
import 'price_history_widget.dart';
import 'pricing_explanation_sheet.dart';
import 'promotion_banner.dart';
import 'surge_indicator.dart';

/// Trip price summary card: original price, discounted price, savings,
/// applied voucher, promotion reason, and a competitor-price badge — each
/// shown only when there is a real value backing it (see each field's doc
/// comment). Tap "Chi tiết" to open the full itemised [PriceBreakdownSheet].
///
/// [breakdown] is a client-side estimate (see [MockFareBreakdown]) — the
/// booking screen has no reachable backend endpoint for fare estimation
/// (`backend/services/pricing`'s `EstimateFare` RPC has no REST gateway
/// route), so this mirrors the real formula locally rather than fabricating
/// unrelated numbers. It is always labelled "ước tính" (estimate), never
/// presented as a confirmed/final price.
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

  final MockFareBreakdown breakdown;

  /// Real trip geometry (haversine distance / estimated duration — see
  /// `MockTripMetrics`), used only to render the "Tại sao giá này?"
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

  bool get _hasDiscount => breakdown.discountCents > 0;
  int get _originalCents => breakdown.totalCents + breakdown.discountCents;

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
                          if (_hasDiscount)
                            Text(
                              breakdown.format(_originalCents),
                              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                                    color: AppColors.textTertiary,
                                    decoration: TextDecoration.lineThrough,
                                  ),
                            ),
                          AnimatedCounter(
                            value: breakdown.totalCents,
                            format: breakdown.format,
                            style: Theme.of(context).textTheme.headlineSmall?.copyWith(color: AppColors.primary),
                          ),
                          Text('Giá ước tính', style: Theme.of(context).textTheme.labelSmall),
                        ],
                      ),
                    ),
                    if (_hasDiscount)
                      AppStatusChip(
                        label: 'Tiết kiệm ${breakdown.format(breakdown.discountCents)}',
                        color: AppColors.primary,
                        icon: Icons.savings_outlined,
                      )
                    else if (cheaperThanCompetitorLabel != null)
                      AppStatusChip(
                        label: cheaperThanCompetitorLabel!,
                        color: AppColors.primary,
                        icon: Icons.trending_down,
                      ),
                  ],
                ),
                if (_hasDiscount) ...[
                  const SizedBox(height: AppSpacing.sm),
                  PriceHistoryWidget(
                    originalCents: _originalCents,
                    finalCents: breakdown.totalCents,
                    format: breakdown.format,
                  ),
                ],
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
