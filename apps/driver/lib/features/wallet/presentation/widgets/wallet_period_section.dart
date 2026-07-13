import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/utils/currency_format.dart';
import '../../../../shared/widgets/charts/mini_line_chart.dart';
import '../../domain/models/wallet_period_totals.dart';

/// Phần 7/13 — "Card: Hôm nay, Tuần, Tháng" + "Biểu đồ" (a simple line
/// chart — reuses the existing `MiniLineChart` `CustomPainter` sparkline,
/// no new chart package, per Phần 7's own "Không package mới").
class WalletPeriodSection extends StatelessWidget {
  const WalletPeriodSection({super.key, required this.totals, required this.dailySeries, required this.currency});

  final WalletPeriodTotals totals;
  final List<int> dailySeries;
  final String currency;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(child: _PeriodCard(label: 'Hôm nay', amountCents: totals.todayCents, currency: currency)),
            const SizedBox(width: AppSpacing.sm),
            Expanded(child: _PeriodCard(label: 'Tuần này', amountCents: totals.weekCents, currency: currency)),
            const SizedBox(width: AppSpacing.sm),
            Expanded(child: _PeriodCard(label: 'Tháng này', amountCents: totals.monthCents, currency: currency)),
          ],
        ),
        const SizedBox(height: AppSpacing.lg),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(AppSpacing.lg),
          decoration: BoxDecoration(
            color: AppColors.surface,
            borderRadius: AppRadius.lgAll,
            border: Border.all(color: AppColors.border),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Xu hướng 7 ngày qua', style: Theme.of(context).textTheme.titleSmall),
              const SizedBox(height: AppSpacing.md),
              MiniLineChart(values: dailySeries, height: 56),
            ],
          ),
        ),
      ],
    );
  }
}

class _PeriodCard extends StatelessWidget {
  const _PeriodCard({required this.label, required this.amountCents, required this.currency});

  final String label;
  final int amountCents;
  final String currency;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.sm, vertical: AppSpacing.md),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: AppRadius.mdAll,
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
          const SizedBox(height: 4),
          Text(
            formatMoney(amountCents, currency),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w800),
          ),
        ],
      ),
    );
  }
}
