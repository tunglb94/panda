import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_card.dart';
import 'package:rider/shared/widgets/app_status_chip.dart';

import '../../domain/models/voucher.dart';

/// One voucher — icon, discount, condition, expiry, budget progress, and a
/// status badge. Selectable only when [Voucher.status] is `available` or
/// `applied`; unavailable/used/expired vouchers render dimmed and inert.
class VoucherCard extends StatelessWidget {
  const VoucherCard({super.key, required this.voucher, this.onTap});

  final Voucher voucher;
  final VoidCallback? onTap;

  bool get _isApplied => voucher.status == VoucherStatus.applied;
  bool get _isDimmed =>
      voucher.status == VoucherStatus.unavailable ||
      voucher.status == VoucherStatus.used ||
      voucher.status == VoucherStatus.expired;

  Color get _statusColor => switch (voucher.status) {
        VoucherStatus.available => AppColors.primary,
        VoucherStatus.applied => AppColors.primary,
        VoucherStatus.unavailable => AppColors.textTertiary,
        VoucherStatus.used => AppColors.textTertiary,
        VoucherStatus.expired => AppColors.error,
      };

  @override
  Widget build(BuildContext context) {
    final iconColor = _isDimmed ? AppColors.textTertiary : voucher.accentColor;

    return Opacity(
      opacity: _isDimmed ? 0.55 : 1,
      child: AppCard(
        onTap: voucher.status.isSelectable ? onTap : null,
        color: _isApplied ? AppColors.primaryLight : AppColors.surface,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(
                    color: iconColor.withValues(alpha: 0.12),
                    borderRadius: AppRadius.mdAll,
                  ),
                  alignment: Alignment.center,
                  child: Icon(voucher.icon, size: AppIconSize.lg, color: iconColor),
                ),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Expanded(
                            child: Text(
                              voucher.title,
                              style: Theme.of(context).textTheme.titleSmall,
                            ),
                          ),
                          Text(
                            voucher.discountLabel,
                            style: Theme.of(context).textTheme.titleSmall?.copyWith(
                                  color: _isDimmed ? AppColors.textTertiary : AppColors.primary,
                                ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 2),
                      Text(
                        voucher.description,
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                  ),
                ),
              ],
            ),
            if (_isApplied) ...[
              const SizedBox(height: AppSpacing.sm),
              Row(
                children: [
                  const Icon(Icons.check_circle, size: AppIconSize.sm, color: AppColors.primary),
                  const SizedBox(width: 6),
                  Expanded(
                    child: Text(
                      'Đã áp dụng · ${voucher.code} · ${voucher.discountLabel}'
                      '${voucher.conditionText != null ? '\nLý do: ${voucher.conditionText}' : ''}',
                      style: Theme.of(context).textTheme.labelMedium?.copyWith(color: AppColors.primary),
                    ),
                  ),
                ],
              ),
            ] else if (voucher.status == VoucherStatus.unavailable && voucher.conditionText != null) ...[
              const SizedBox(height: AppSpacing.sm),
              Row(
                children: [
                  Icon(Icons.cancel_outlined, size: AppIconSize.sm, color: AppColors.textTertiary),
                  const SizedBox(width: 6),
                  Expanded(
                    child: Text(
                      'Không đủ điều kiện · ${voucher.conditionText}',
                      style: Theme.of(context).textTheme.labelMedium?.copyWith(color: AppColors.textTertiary),
                    ),
                  ),
                ],
              ),
            ] else if (voucher.conditionText != null) ...[
              const SizedBox(height: AppSpacing.sm),
              Row(
                children: [
                  Icon(Icons.info_outline, size: AppIconSize.sm, color: AppColors.textTertiary),
                  const SizedBox(width: 6),
                  Expanded(
                    child: Text(
                      voucher.conditionText!,
                      style: Theme.of(context).textTheme.labelMedium,
                    ),
                  ),
                ],
              ),
            ],
            if (voucher.budgetUsedRatio != null) ...[
              const SizedBox(height: AppSpacing.sm),
              ClipRRect(
                borderRadius: AppRadius.smAll,
                child: LinearProgressIndicator(
                  value: voucher.budgetUsedRatio!.clamp(0, 1),
                  minHeight: 6,
                  backgroundColor: AppColors.divider,
                  valueColor: AlwaysStoppedAnimation(
                    voucher.budgetUsedRatio! >= 0.9 ? AppColors.warning : AppColors.primary,
                  ),
                ),
              ),
              const SizedBox(height: 4),
              Text(
                'Đã dùng ${(voucher.budgetUsedRatio! * 100).round()}% ngân sách khuyến mãi',
                style: Theme.of(context).textTheme.labelSmall,
              ),
            ],
            const SizedBox(height: AppSpacing.sm),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                if (voucher.expiresAt != null)
                  Text(
                    'HSD: ${_formatDate(voucher.expiresAt!)}',
                    style: Theme.of(context).textTheme.labelSmall,
                  )
                else
                  const SizedBox.shrink(),
                AppStatusChip(
                  label: voucher.status.badgeLabel,
                  color: _statusColor,
                  icon: _isApplied ? Icons.check_circle : null,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  String _formatDate(DateTime d) =>
      '${d.day.toString().padLeft(2, '0')}/${d.month.toString().padLeft(2, '0')}/${d.year}';
}
