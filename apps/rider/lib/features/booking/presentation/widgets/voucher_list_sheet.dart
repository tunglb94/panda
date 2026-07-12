import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_bottom_sheet.dart';
import 'package:rider/shared/widgets/app_button.dart';
import 'package:rider/shared/widgets/app_empty_state.dart';

import '../../domain/models/voucher.dart';
import '../../domain/models/voucher_catalog.dart';
import 'voucher_card.dart';

/// Opens the voucher picker sheet. Returns the tapped [Voucher], or `null`
/// if the rider dismissed the sheet or chose "Bỏ chọn voucher".
abstract final class VoucherListSheet {
  static Future<Voucher?> show(BuildContext context, {Voucher? selected}) {
    return AppBottomSheet.show<Voucher?>(
      context,
      title: 'Chọn voucher',
      isScrollControlled: true,
      builder: (sheetContext) => _VoucherListBody(selected: selected),
    );
  }
}

class _VoucherListBody extends StatelessWidget {
  const _VoucherListBody({this.selected});

  final Voucher? selected;

  @override
  Widget build(BuildContext context) {
    final vouchers = VoucherCatalog.mine;

    if (vouchers.isEmpty) {
      return SizedBox(
        height: 360,
        child: AppEmptyState(
          icon: Icons.local_offer_outlined,
          title: 'Chưa có voucher nào',
          subtitle: 'Ưu đãi mới sẽ xuất hiện ở đây ngay khi có sẵn cho bạn.',
          mascotAsset: 'mascot_voucher.png',
        ),
      );
    }

    return ConstrainedBox(
      constraints: BoxConstraints(maxHeight: MediaQuery.of(context).size.height * 0.7),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Flexible(
            child: ListView.separated(
              shrinkWrap: true,
              itemCount: vouchers.length,
              separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.md),
              itemBuilder: (context, i) {
                final voucher = vouchers[i];
                final isSelected = voucher.id == selected?.id;
                return VoucherCard(
                  voucher: isSelected
                      ? Voucher(
                          id: voucher.id,
                          code: voucher.code,
                          title: voucher.title,
                          description: voucher.description,
                          icon: voucher.icon,
                          accentColor: voucher.accentColor,
                          discountLabel: voucher.discountLabel,
                          status: VoucherStatus.applied,
                          conditionText: voucher.conditionText,
                          expiresAt: voucher.expiresAt,
                          budgetUsedRatio: voucher.budgetUsedRatio,
                          discountPercent: voucher.discountPercent,
                        )
                      : voucher,
                  onTap: () => Navigator.of(context).pop(voucher),
                );
              },
            ),
          ),
          if (selected != null) ...[
            const SizedBox(height: AppSpacing.md),
            AppButton.text(
              label: 'Bỏ chọn voucher',
              icon: Icons.close,
              onPressed: () => Navigator.of(context).pop(),
            ),
          ],
        ],
      ),
    );
  }
}

/// Tappable row shown in the booking form: "Chọn voucher" when nothing is
/// selected, or the applied voucher's title + discount once one is chosen.
class VoucherPickerTile extends StatelessWidget {
  const VoucherPickerTile({super.key, required this.selected, required this.onTap});

  final Voucher? selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final voucher = selected;
    return InkWell(
      onTap: onTap,
      borderRadius: AppRadius.mdAll,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
        child: Row(
          children: [
            Icon(
              Icons.local_offer_outlined,
              size: AppIconSize.md,
              color: voucher != null ? AppColors.primary : AppColors.textSecondary,
            ),
            const SizedBox(width: AppSpacing.sm),
            Expanded(
              child: Text(
                voucher != null ? '${voucher.title} · ${voucher.discountLabel}' : 'Chọn voucher',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: voucher != null ? AppColors.textPrimary : AppColors.textSecondary,
                      fontWeight: voucher != null ? FontWeight.w600 : FontWeight.w400,
                    ),
              ),
            ),
            Icon(Icons.chevron_right, size: AppIconSize.md, color: AppColors.textTertiary),
          ],
        ),
      ),
    );
  }
}
