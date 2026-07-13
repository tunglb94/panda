import 'package:flutter/material.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/utils/currency_format.dart';
import '../../../wallet/data/wallet_repository.dart';
import '../../../wallet/domain/models/wallet_summary.dart';
import '../../../wallet/presentation/pages/wallet_page.dart';

/// Wallet summary preview, tapping through to the full [WalletPage] (Phần
/// 13). Backed by the real Settlement Engine (`WalletRepository`) — no
/// fabricated numbers; a fetch failure just shows "—" rather than a fake
/// balance, exactly like this card's pre-Financial-Core placeholder did.
class WalletCard extends StatefulWidget {
  const WalletCard({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<WalletCard> createState() => _WalletCardState();
}

class _WalletCardState extends State<WalletCard> {
  late final WalletRepository _repo = WalletRepository(widget.apiClient);
  WalletSummary? _summary;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    try {
      final summary = await _repo.fetchSummary();
      if (mounted) setState(() => _summary = summary);
    } catch (_) {
      // Leave as "—" — see class doc.
    }
  }

  void _openWallet() {
    Navigator.of(context)
        .push(MaterialPageRoute(builder: (_) => WalletPage(apiClient: widget.apiClient)))
        .then((_) => _load());
  }

  @override
  Widget build(BuildContext context) {
    final summary = _summary;
    return InkWell(
      onTap: _openWallet,
      borderRadius: AppRadius.lgAll,
      child: Container(
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
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(color: Colors.white),
                    ),
                  ],
                ),
                const Icon(Icons.chevron_right, color: Colors.white),
              ],
            ),
            const SizedBox(height: AppSpacing.lg),
            Text(
              'Số dư khả dụng',
              style: TextStyle(color: Colors.white.withValues(alpha: 0.8), fontSize: 12),
            ),
            const SizedBox(height: 4),
            Text(
              summary == null ? '—' : formatMoney(summary.availableCents, summary.currency),
              style: const TextStyle(color: Colors.white, fontSize: 32, fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: AppSpacing.lg),
            Row(
              children: [
                Expanded(
                  child: _BalanceSubStat(
                    label: 'Đang chờ',
                    value: summary == null ? '—' : formatMoney(summary.pendingCents, summary.currency),
                  ),
                ),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: _BalanceSubStat(
                    label: 'Đang nợ Panda',
                    value: summary == null ? '—' : formatMoney(summary.outstandingCents, summary.currency),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _BalanceSubStat extends StatelessWidget {
  const _BalanceSubStat({required this.label, required this.value});

  final String label;
  final String value;

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
          Text(
            value,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w700),
          ),
        ],
      ),
    );
  }
}
