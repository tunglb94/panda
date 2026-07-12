import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_status_chip.dart';

/// Wallet section. There is no wallet service in the backend today (no
/// balance, no withdrawal, no payout-method storage — see
/// `docs/driver/DRIVER_APP_SPEC.md` §10) and this task adds none. Every
/// figure here is an honest "—" placeholder inside a premium-styled shell,
/// with a visible "Sắp ra mắt" badge — never a fabricated balance. The
/// visual design (gradient card, sub-balance breakdown, payment-method
/// chips) is built now so wiring real data later is a data-layer change,
/// not a redesign.
class WalletCard extends StatelessWidget {
  const WalletCard({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.xl),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [AppColors.primary, AppColors.primaryDark],
        ),
        borderRadius: AppRadius.lgAll,
        boxShadow: [
          BoxShadow(
            color: AppColors.primary.withValues(alpha: 0.28),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  const Icon(Icons.account_balance_wallet, color: Colors.white, size: AppIconSize.lg),
                  const SizedBox(width: AppSpacing.sm),
                  Text(
                    'Ví Panda',
                    style: Theme.of(context)
                        .textTheme
                        .titleMedium
                        ?.copyWith(color: Colors.white),
                  ),
                ],
              ),
              const AppStatusChip(label: 'Sắp ra mắt', color: Colors.white),
            ],
          ),
          const SizedBox(height: AppSpacing.lg),
          Text(
            'Số dư khả dụng',
            style: TextStyle(color: Colors.white.withValues(alpha: 0.8), fontSize: 12),
          ),
          const SizedBox(height: 4),
          const Text(
            '—',
            style: TextStyle(color: Colors.white, fontSize: 32, fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: AppSpacing.lg),
          Row(
            children: const [
              Expanded(child: _BalanceSubStat(label: 'Đang chờ')),
              SizedBox(width: AppSpacing.md),
              Expanded(child: _BalanceSubStat(label: 'Bị tạm giữ')),
            ],
          ),
          const SizedBox(height: AppSpacing.lg),
          Wrap(
            spacing: AppSpacing.sm,
            runSpacing: AppSpacing.sm,
            children: const [
              _PaymentMethodChip(icon: Icons.payments_outlined, label: 'Tiền mặt'),
              _PaymentMethodChip(icon: Icons.account_balance_outlined, label: 'Ngân hàng'),
              _PaymentMethodChip(icon: Icons.card_giftcard_outlined, label: 'Voucher'),
              _PaymentMethodChip(icon: Icons.emoji_events_outlined, label: 'Thưởng'),
              _PaymentMethodChip(icon: Icons.monetization_on_outlined, label: 'Xu'),
            ],
          ),
        ],
      ),
    );
  }
}

class _BalanceSubStat extends StatelessWidget {
  const _BalanceSubStat({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.12),
        borderRadius: AppRadius.smAll,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: TextStyle(color: Colors.white.withValues(alpha: 0.75), fontSize: 11)),
          const SizedBox(height: 2),
          const Text('—', style: TextStyle(color: Colors.white, fontWeight: FontWeight.w700)),
        ],
      ),
    );
  }
}

class _PaymentMethodChip extends StatelessWidget {
  const _PaymentMethodChip({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.12),
        borderRadius: AppRadius.pillAll,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: Colors.white),
          const SizedBox(width: 5),
          Text(label, style: const TextStyle(color: Colors.white, fontSize: 11, fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }
}
