import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_card.dart';
import 'package:rider/shared/widgets/app_empty_state.dart';
import 'package:rider/shared/widgets/app_snackbar.dart';

/// Wallet tab — new for the Closed Beta polish pass (PHẦN 8). There is no
/// wallet/payment/voucher backend anywhere in this project (confirmed via
/// audit: no `/wallet`, `/vouchers`, or `/promotions` endpoints exist, and
/// `PaymentMethodCard`'s selection is a client-side-only mock). Every
/// number and list here is therefore an honest, clearly-labeled
/// placeholder — no invented balance, no fabricated voucher catalog —
/// exactly mirroring how `apps/driver`'s Earnings/Wallet sprint handled the
/// same situation before its wallet endpoints existed.
class WalletPage extends StatelessWidget {
  const WalletPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Ví')),
      body: ListView(
        padding: const EdgeInsets.all(AppSpacing.lg),
        children: [
          const _BalanceCard(),
          const SizedBox(height: AppSpacing.lg),
          const _QuickActionsRow(),
          const SizedBox(height: AppSpacing.xxl),
          Text('Ưu đãi & Voucher', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: AppSpacing.md),
          const AppEmptyState(
            icon: Icons.local_offer_outlined,
            title: 'Chưa có ưu đãi nào',
            subtitle: 'Các mã giảm giá và voucher khả dụng sẽ hiển thị ở đây.',
            mascotAsset: 'mascot_voucher.png',
          ),
          const SizedBox(height: AppSpacing.xxl),
          Text('Lịch sử giao dịch', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: AppSpacing.md),
          AppEmptyState(
            icon: Icons.receipt_long_outlined,
            title: 'Chưa khả dụng',
            subtitle: 'Lịch sử nạp/rút và thanh toán qua ví sẽ ra mắt trong giai đoạn tiếp theo.',
          ),
        ],
      ),
    );
  }
}

class _BalanceCard extends StatelessWidget {
  const _BalanceCard();

  @override
  Widget build(BuildContext context) {
    return AppCard(
      color: AppColors.primary,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.account_balance_wallet_outlined, color: AppColors.textOnPrimary, size: AppIconSize.lg),
              const SizedBox(width: AppSpacing.sm),
              Text(
                'Số dư ví',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: AppColors.textOnPrimary),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.sm),
          Text(
            '—',
            style: Theme.of(context).textTheme.headlineMedium?.copyWith(color: AppColors.textOnPrimary),
          ),
          const SizedBox(height: 2),
          Text(
            'Ví Panda chưa khả dụng — sẽ ra mắt trong giai đoạn tiếp theo.',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: AppColors.textOnPrimary.withValues(alpha: 0.85),
                ),
          ),
        ],
      ),
    );
  }
}

class _QuickActionsRow extends StatelessWidget {
  const _QuickActionsRow();

  @override
  Widget build(BuildContext context) {
    final actions = <(IconData, String)>[
      (Icons.add_circle_outline, 'Nạp tiền'),
      (Icons.arrow_circle_up_outlined, 'Rút tiền'),
      (Icons.card_giftcard_outlined, 'Mã ưu đãi'),
    ];
    return Row(
      children: [
        for (final (icon, label) in actions) ...[
          Expanded(child: _QuickActionButton(icon: icon, label: label)),
          if (label != actions.last.$2) const SizedBox(width: AppSpacing.sm),
        ],
      ],
    );
  }
}

class _QuickActionButton extends StatelessWidget {
  const _QuickActionButton({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: AppColors.surface,
      borderRadius: AppRadius.mdAll,
      child: InkWell(
        borderRadius: AppRadius.mdAll,
        onTap: () => AppSnackbar.show(context, '$label chưa khả dụng — sẽ ra mắt trong giai đoạn tiếp theo.'),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: AppSpacing.md),
          decoration: BoxDecoration(
            borderRadius: AppRadius.mdAll,
            border: Border.all(color: AppColors.border),
          ),
          child: Column(
            children: [
              Icon(icon, color: AppColors.primary, size: AppIconSize.lg),
              const SizedBox(height: 6),
              Text(
                label,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.labelMedium,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
