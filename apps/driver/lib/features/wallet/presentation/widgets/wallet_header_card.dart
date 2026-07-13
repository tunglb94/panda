import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/utils/currency_format.dart';
import '../../domain/models/wallet_summary.dart';

/// Phần 13's Wallet screen header — Available balance (large), Pending
/// balance, an Outstanding warning banner when a driver owes Panda
/// commission from cash trips (Phần 4), and the "Rút tiền" button, disabled
/// with a Vietnamese tooltip whenever Phần 5's conditions aren't met yet
/// (mirrors the Driver KYC phase's disabled-button-with-Tooltip pattern).
class WalletHeaderCard extends StatelessWidget {
  const WalletHeaderCard({
    super.key,
    required this.summary,
    required this.canWithdraw,
    required this.withdrawBlockedReason,
    required this.onWithdraw,
    required this.onTapBankAccount,
    this.bankAccountLabel,
  });

  final WalletSummary summary;
  final bool canWithdraw;
  final String? withdrawBlockedReason;
  final VoidCallback onWithdraw;
  final VoidCallback onTapBankAccount;
  final String? bankAccountLabel;

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
          BoxShadow(color: AppColors.primary.withValues(alpha: 0.28), blurRadius: 20, offset: const Offset(0, 8)),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.account_balance_wallet, color: Colors.white, size: AppIconSize.lg),
              const SizedBox(width: AppSpacing.sm),
              Text('Ví Panda', style: Theme.of(context).textTheme.titleMedium?.copyWith(color: Colors.white)),
            ],
          ),
          const SizedBox(height: AppSpacing.lg),
          Text('Số dư khả dụng', style: TextStyle(color: Colors.white.withValues(alpha: 0.8), fontSize: 12)),
          const SizedBox(height: 4),
          Text(
            formatMoney(summary.availableCents, summary.currency),
            style: const TextStyle(color: Colors.white, fontSize: 32, fontWeight: FontWeight.w800),
          ),
          const SizedBox(height: AppSpacing.lg),
          Row(
            children: [
              Expanded(child: _SubStat(label: 'Đang chờ', value: formatMoney(summary.pendingCents, summary.currency))),
              const SizedBox(width: AppSpacing.md),
              Expanded(
                child: InkWell(
                  onTap: onTapBankAccount,
                  borderRadius: AppRadius.smAll,
                  child: _SubStat(
                    label: 'Ngân hàng',
                    value: bankAccountLabel ?? 'Thêm ngân hàng',
                    trailingIcon: Icons.chevron_right,
                  ),
                ),
              ),
            ],
          ),
          if (summary.outstandingCents > 0) ...[
            const SizedBox(height: AppSpacing.md),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(AppSpacing.sm),
              decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.16), borderRadius: AppRadius.smAll),
              child: Row(
                children: [
                  const Icon(Icons.error_outline, color: Colors.white, size: 16),
                  const SizedBox(width: 6),
                  Expanded(
                    child: Text(
                      'Bạn đang nợ Panda ${formatMoney(summary.outstandingCents, summary.currency)} tiền hoa hồng từ chuyến thu tiền mặt.',
                      style: const TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.w600),
                    ),
                  ),
                ],
              ),
            ),
          ],
          const SizedBox(height: AppSpacing.lg),
          Tooltip(
            message: canWithdraw ? 'Rút tiền về ngân hàng' : (withdrawBlockedReason ?? 'Chưa thể rút tiền'),
            child: Semantics(
              label: canWithdraw ? null : 'Nút rút tiền bị khoá — ${withdrawBlockedReason ?? "chưa thể rút tiền"}',
              child: SizedBox(
                width: double.infinity,
                child: FilledButton.icon(
                  onPressed: canWithdraw ? onWithdraw : null,
                  style: FilledButton.styleFrom(
                    backgroundColor: Colors.white,
                    foregroundColor: AppColors.primary,
                    disabledBackgroundColor: Colors.white.withValues(alpha: 0.4),
                    disabledForegroundColor: AppColors.primary.withValues(alpha: 0.6),
                  ),
                  icon: const Icon(Icons.arrow_circle_up_outlined),
                  label: const Text('Rút tiền', style: TextStyle(fontWeight: FontWeight.w700)),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _SubStat extends StatelessWidget {
  const _SubStat({required this.label, required this.value, this.trailingIcon});

  final String label;
  final String value;
  final IconData? trailingIcon;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
      decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.12), borderRadius: AppRadius.smAll),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: TextStyle(color: Colors.white.withValues(alpha: 0.75), fontSize: 11)),
          const SizedBox(height: 2),
          Row(
            children: [
              Expanded(
                child: Text(
                  value,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w700, fontSize: 13),
                ),
              ),
              if (trailingIcon != null) Icon(trailingIcon, color: Colors.white, size: 16),
            ],
          ),
        ],
      ),
    );
  }
}
