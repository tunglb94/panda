import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_bottom_sheet.dart';
import 'package:rider/shared/widgets/app_button.dart';
import 'package:rider/shared/widgets/app_empty_state.dart';

import '../../data/promotion_repository.dart';
import '../../domain/models/voucher.dart';
import 'voucher_card.dart';

/// Opens the voucher picker sheet — real vouchers from
/// `GET /api/v1/rider/vouchers`'s "available" list. Returns the tapped
/// [Voucher], or `null` if the rider dismissed the sheet or chose "Bỏ chọn
/// voucher".
abstract final class VoucherListSheet {
  static Future<Voucher?> show(BuildContext context, {required ApiClient apiClient, Voucher? selected}) {
    return AppBottomSheet.show<Voucher?>(
      context,
      title: 'Chọn voucher',
      isScrollControlled: true,
      builder: (sheetContext) => _VoucherListBody(apiClient: apiClient, selected: selected),
    );
  }
}

class _VoucherListBody extends StatefulWidget {
  const _VoucherListBody({required this.apiClient, this.selected});

  final ApiClient apiClient;
  final Voucher? selected;

  @override
  State<_VoucherListBody> createState() => _VoucherListBodyState();
}

class _VoucherListBodyState extends State<_VoucherListBody> {
  bool _loading = true;
  List<Voucher> _vouchers = const [];

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final body = await PromotionRepository(widget.apiClient).myVouchers();
      final available = (body['available'] as List<dynamic>? ?? [])
          .map((e) => Voucher.fromApi(e as Map<String, dynamic>, status: VoucherStatus.available))
          .toList();
      if (!mounted) return;
      setState(() {
        _vouchers = available;
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return const SizedBox(height: 200, child: Center(child: CircularProgressIndicator()));
    }

    if (_vouchers.isEmpty) {
      return const SizedBox(
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
              itemCount: _vouchers.length,
              separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.md),
              itemBuilder: (context, i) {
                final voucher = _vouchers[i];
                final isSelected = voucher.id == widget.selected?.id;
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
          if (widget.selected != null) ...[
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
