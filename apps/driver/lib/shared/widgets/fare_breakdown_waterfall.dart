import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_icon_sizes.dart';
import '../../core/theme/app_spacing.dart';
import '../utils/currency_format.dart';
import 'animated_counter.dart';
import 'app_card.dart';

/// "Khách trả → Voucher → Platform Fee → Commission → VAT → Thu nhập tài
/// xế" waterfall — Driver Earnings, Payment/Fare production pass. One
/// widget shared by the Earnings tab (period summary) and a trip's own
/// award/completion card (single-trip amount) — no duplicated logic.
///
/// Only [grossAmountCents] is ever real (the sum/single value of
/// `final_fare`, the fare the rider actually paid — the only money field
/// `TripProto`/any gateway response exposes). The backend has no field
/// anywhere for voucher subsidy, platform fee, commission, or VAT split
/// (see `FullFareBreakdownV3` in `backend/services/pricing`, which computes
/// all of this internally but never crosses the gRPC boundary), so those
/// rows always show "Đang cập nhật" rather than a guessed split. This
/// widget deliberately never repeats [grossAmountCents] under the "Thu
/// nhập tài xế" label — BRB's commission tiers (§7.1) mean a driver's real
/// net income is always less than the gross fare, so showing the same
/// number under two labels would misstate take-home pay.
class FareBreakdownWaterfall extends StatelessWidget {
  const FareBreakdownWaterfall({
    super.key,
    required this.grossAmountCents,
    required this.currency,
    this.title = 'Chi tiết thu nhập',
  });

  final int grossAmountCents;
  final String currency;
  final String title;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.account_balance_wallet_outlined,
                  size: AppIconSize.md, color: AppColors.primary),
              const SizedBox(width: AppSpacing.sm),
              Text(title, style: Theme.of(context).textTheme.titleSmall),
            ],
          ),
          const SizedBox(height: AppSpacing.md),
          _row(
            context,
            label: 'Khách trả',
            trailing: grossAmountCents > 0 && currency.isNotEmpty
                ? AnimatedCounter(
                    value: grossAmountCents,
                    format: (v) => formatMoney(v, currency),
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w700),
                  )
                : null,
          ),
          _arrow(),
          _row(context, label: 'Voucher'),
          _arrow(),
          _row(context, label: 'Platform Fee'),
          _arrow(),
          _row(context, label: 'Commission'),
          _arrow(),
          _row(context, label: 'VAT'),
          _arrow(),
          _row(context, label: 'Thu nhập tài xế', isFinal: true),
        ],
      ),
    );
  }

  Widget _row(BuildContext context, {required String label, Widget? trailing, bool isFinal = false}) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: isFinal
                ? theme.textTheme.titleSmall
                : theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
          ),
          Flexible(
            child: Align(
              alignment: Alignment.centerRight,
              child: trailing ??
                  Text(
                    'Đang cập nhật',
                    style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textTertiary),
                  ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _arrow() {
    return const Icon(Icons.arrow_downward, size: AppIconSize.sm, color: AppColors.textTertiary);
  }
}
