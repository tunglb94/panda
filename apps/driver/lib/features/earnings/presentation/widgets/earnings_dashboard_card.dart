import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_shadows.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/utils/currency_format.dart';
import '../../../../shared/widgets/animated_counter.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_status_chip.dart';
import '../../../../shared/widgets/charts/mini_bar_chart.dart';
import '../../domain/models/earnings_models.dart';

/// The Earnings Dashboard: period tabs (Ngày/Tuần/Tháng), the real
/// aggregated total for that period, completed/cancelled trip counts, and a
/// 7-day trend bar chart — all computed from the driver's actual trip
/// history (`GET /api/v1/driver/trips`), never fabricated. The
/// voucher/platform-support/commission breakdown lives in the sibling
/// `EarningsBreakdownCard`, shown directly below this one.
class EarningsDashboardCard extends StatelessWidget {
  const EarningsDashboardCard({
    super.key,
    required this.summary,
    required this.selectedPeriod,
    required this.onPeriodChanged,
  });

  final EarningsSummary summary;
  final EarningsPeriod selectedPeriod;
  final ValueChanged<EarningsPeriod> onPeriodChanged;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      padding: const EdgeInsets.all(AppSpacing.xl),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _PeriodTabs(selected: selectedPeriod, onChanged: onPeriodChanged),
          const SizedBox(height: AppSpacing.lg),
          Text(
            'Tổng thu nhập',
            style: Theme.of(context).textTheme.bodySmall,
          ),
          const SizedBox(height: AppSpacing.xs),
          AnimatedCounter(
            value: summary.totalCents,
            format: (v) => formatMoney(v, summary.currency),
            style: Theme.of(context)
                .textTheme
                .headlineMedium
                ?.copyWith(color: AppColors.primary),
          ),
          const SizedBox(height: AppSpacing.lg),
          Wrap(
            spacing: AppSpacing.sm,
            runSpacing: AppSpacing.sm,
            children: [
              AppStatusChip(
                icon: Icons.check_circle_outline,
                label: '${summary.completedCount} hoàn thành',
                color: AppColors.primary,
              ),
              AppStatusChip(
                icon: Icons.cancel_outlined,
                label: '${summary.cancelledCount} đã hủy',
                color: AppColors.textTertiary,
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.lg),
          Text(
            'Xu hướng 7 ngày qua',
            style: Theme.of(context).textTheme.labelMedium,
          ),
          const SizedBox(height: AppSpacing.sm),
          MiniBarChart(
            values: summary.dailySeries.map((p) => p.amountCents).toList(),
            labels: summary.dailySeries.map((p) => _weekdayLabel(p.day)).toList(),
          ),
        ],
      ),
    );
  }

  static String _weekdayLabel(DateTime d) {
    const labels = ['T2', 'T3', 'T4', 'T5', 'T6', 'T7', 'CN'];
    return labels[d.weekday - 1];
  }
}

class _PeriodTabs extends StatelessWidget {
  const _PeriodTabs({required this.selected, required this.onChanged});

  final EarningsPeriod selected;
  final ValueChanged<EarningsPeriod> onChanged;

  static const _options = [
    (EarningsPeriod.day, 'Hôm nay'),
    (EarningsPeriod.week, 'Tuần này'),
    (EarningsPeriod.month, 'Tháng này'),
  ];

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.xs),
      decoration: BoxDecoration(
        color: AppColors.surfaceAlt,
        borderRadius: AppRadius.mdAll,
      ),
      child: Row(
        children: _options.map((o) {
          final isSelected = o.$1 == selected;
          return Expanded(
            child: GestureDetector(
              onTap: () => onChanged(o.$1),
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                padding: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
                decoration: BoxDecoration(
                  color: isSelected ? AppColors.surface : Colors.transparent,
                  borderRadius: AppRadius.smAll,
                  boxShadow: isSelected ? AppShadows.card : AppShadows.none,
                ),
                alignment: Alignment.center,
                child: Text(
                  o.$2,
                  style: Theme.of(context).textTheme.labelMedium?.copyWith(
                        fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                        color: isSelected ? AppColors.primary : AppColors.textSecondary,
                      ),
                ),
              ),
            ),
          );
        }).toList(),
      ),
    );
  }
}
