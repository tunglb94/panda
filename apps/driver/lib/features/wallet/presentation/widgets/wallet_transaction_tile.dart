import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/utils/currency_format.dart';
import '../../domain/models/wallet_transaction.dart';

/// One row of the Phần 6/9 Transaction History — Ride/Delivery/Commission/
/// Bonus/Withdrawal/Refund/Adjustment, tagged Cash/Electronic when relevant.
class WalletTransactionTile extends StatelessWidget {
  const WalletTransactionTile({super.key, required this.transaction});

  final WalletTransaction transaction;

  IconData get _icon => switch (transaction.type) {
        'ride_income' => Icons.two_wheeler,
        'delivery_income' => Icons.local_shipping_outlined,
        'commission' || 'platform_receivable' => Icons.percent,
        'platform_payable' => Icons.account_balance_outlined,
        'promotion_subsidy' || 'voucher_subsidy' => Icons.card_giftcard_outlined,
        'bonus' => Icons.emoji_events_outlined,
        'penalty' => Icons.report_gmailerrorred_outlined,
        'withdrawal' => Icons.arrow_circle_up_outlined,
        'refund' => Icons.replay_outlined,
        _ => Icons.receipt_long_outlined,
      };

  String _dateLabel(DateTime? d) {
    if (d == null) return '';
    final local = d.toLocal();
    return '${local.day.toString().padLeft(2, '0')}/${local.month.toString().padLeft(2, '0')}/${local.year} '
        '${local.hour.toString().padLeft(2, '0')}:${local.minute.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    final color = transaction.isCredit ? AppColors.primary : AppColors.error;
    final sign = transaction.isCredit ? '+' : '-';
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(color: color.withValues(alpha: 0.1), shape: BoxShape.circle),
            child: Icon(_icon, color: color, size: 18),
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
                        walletTransactionTypeLabel(transaction.type),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: Theme.of(context).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
                      ),
                    ),
                    if (transaction.paymentMethod.isNotEmpty)
                      Padding(
                        padding: const EdgeInsets.only(left: 6),
                        child: Text(
                          transaction.paymentMethod == 'cash' ? 'Tiền mặt' : 'Điện tử',
                          style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textTertiary),
                        ),
                      ),
                  ],
                ),
                const SizedBox(height: 2),
                Text(
                  _dateLabel(transaction.createdAt),
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
                ),
              ],
            ),
          ),
          Text(
            '$sign${formatMoney(transaction.amountCents, transaction.currency)}',
            style: TextStyle(color: color, fontWeight: FontWeight.w700),
          ),
        ],
      ),
    );
  }
}
