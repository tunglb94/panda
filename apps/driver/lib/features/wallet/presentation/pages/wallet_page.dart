import 'package:flutter/material.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_bottom_sheet.dart';
import '../../../../shared/widgets/app_empty_state.dart';
import '../../../../shared/widgets/app_loading_view.dart';
import '../../../../shared/widgets/app_snackbar.dart';
import '../../../../shared/widgets/app_status_chip.dart';
import '../../../../shared/utils/currency_format.dart';
import '../../../../core/theme/app_colors.dart';
import '../../data/wallet_repository.dart';
import '../../domain/models/bank_account.dart';
import '../../domain/models/payout_request.dart';
import '../../domain/models/wallet_period_totals.dart';
import '../../domain/models/wallet_summary.dart';
import '../../domain/models/wallet_transaction.dart';
import '../widgets/bank_account_sheet.dart';
import '../widgets/payout_request_sheet.dart';
import '../widgets/wallet_header_card.dart';
import '../widgets/wallet_period_section.dart';
import '../widgets/wallet_transaction_tile.dart';

/// Phần 13's Wallet screen — Header (Available/Pending + Rút tiền) → Card
/// (Hôm nay/Tuần/Tháng) → Biểu đồ → Lịch sử giao dịch. Material 3, dark
/// mode via existing theme tokens, no overflow (bounded widths + ellipsis
/// throughout), Semantics/Tooltip on the Rút tiền button (Phần 13's
/// Accessibility requirement).
class WalletPage extends StatefulWidget {
  const WalletPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<WalletPage> createState() => _WalletPageState();
}

class _WalletPageState extends State<WalletPage> {
  late final WalletRepository _repo = WalletRepository(widget.apiClient);
  bool _loading = true;
  String? _error;
  WalletSummary _summary = WalletSummary.empty;
  List<WalletTransaction> _transactions = [];
  List<PayoutRequest> _payouts = [];
  BankAccount? _bankAccount;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final summary = await _repo.fetchSummary();
      final transactions = await _repo.fetchTransactions();
      BankAccount? bankAccount;
      List<PayoutRequest> payouts = [];
      try {
        bankAccount = await _repo.fetchBankAccount();
      } catch (_) {
        // Best-effort — Wallet still works without a bank account yet.
      }
      try {
        payouts = await _repo.fetchMyPayoutRequests();
      } catch (_) {
        // Best-effort.
      }
      if (mounted) {
        setState(() {
          _summary = summary;
          _transactions = transactions;
          _bankAccount = bankAccount;
          _payouts = payouts;
          _loading = false;
        });
      }
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = e.statusCode == 0 ? e.message : 'Không thể tải dữ liệu ví.';
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = 'Không thể tải dữ liệu ví.';
        });
      }
    }
  }

  bool get _hasInFlightPayout => _payouts.any((p) => p.status == 'pending' || p.status == 'approved');

  bool get _canWithdraw => _summary.netCents > 0 && _bankAccount != null && !_hasInFlightPayout;

  String? get _withdrawBlockedReason {
    if (_bankAccount == null) return 'Cần thêm ngân hàng trước khi rút tiền';
    if (_hasInFlightPayout) return 'Bạn đang có yêu cầu rút tiền chờ xử lý';
    if (_summary.netCents <= 0) return 'Số dư khả dụng không đủ';
    return null;
  }

  Future<void> _openBankAccountSheet() async {
    final saved = await AppBottomSheet.show<bool>(
      context,
      title: null,
      isScrollControlled: true,
      builder: (_) => BankAccountSheet(repository: _repo, existing: _bankAccount),
    );
    if (saved == true) {
      if (mounted) AppSnackbar.success(context, 'Đã lưu tài khoản ngân hàng.');
      _load();
    }
  }

  Future<void> _openPayoutSheet() async {
    final requested = await AppBottomSheet.show<bool>(
      context,
      title: null,
      isScrollControlled: true,
      builder: (_) => PayoutRequestSheet(repository: _repo, summary: _summary),
    );
    if (requested == true) {
      if (mounted) AppSnackbar.success(context, 'Đã gửi yêu cầu rút tiền. Panda sẽ xử lý sớm nhất.');
      _load();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Ví Panda')),
      body: SafeArea(child: _buildBody()),
    );
  }

  Widget _buildBody() {
    if (_loading) return const AppLoadingView(label: 'Đang tải ví…');
    if (_error != null && _transactions.isEmpty) {
      return AppEmptyState.error(
        subtitle: _error!,
        onAction: _load,
        mascotAsset: 'mascot_no_connection.png',
      );
    }

    final bankLabel = _bankAccount == null
        ? null
        : '${_bankAccount!.bankName} ${_bankAccount!.maskedAccountNumber}';
    final totals = WalletPeriodTotals.fromTransactions(_transactions);
    final dailySeries = WalletPeriodTotals.dailySeries(_transactions);

    return RefreshIndicator(
      onRefresh: _load,
      child: ListView(
        padding: const EdgeInsets.all(AppSpacing.lg),
        children: [
          WalletHeaderCard(
            summary: _summary,
            canWithdraw: _canWithdraw,
            withdrawBlockedReason: _withdrawBlockedReason,
            onWithdraw: _openPayoutSheet,
            onTapBankAccount: _openBankAccountSheet,
            bankAccountLabel: bankLabel,
          ),
          const SizedBox(height: AppSpacing.lg),
          WalletPeriodSection(totals: totals, dailySeries: dailySeries, currency: _summary.currency),
          const SizedBox(height: AppSpacing.xl),
          Text('Lịch sử giao dịch', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: AppSpacing.sm),
          if (_transactions.isEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: AppSpacing.lg),
              child: Text(
                'Chưa có giao dịch nào.',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
              ),
            )
          else
            ..._transactions.map((t) => WalletTransactionTile(transaction: t)),
          if (_payouts.isNotEmpty) ...[
            const SizedBox(height: AppSpacing.xl),
            Text('Lịch sử rút tiền', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: AppSpacing.sm),
            ..._payouts.map((p) => _PayoutHistoryRow(payout: p)),
          ],
        ],
      ),
    );
  }
}

class _PayoutHistoryRow extends StatelessWidget {
  const _PayoutHistoryRow({required this.payout});

  final PayoutRequest payout;

  Color get _color => switch (payout.status) {
        'paid' => AppColors.primary,
        'rejected' => AppColors.error,
        'approved' => AppColors.info,
        _ => AppColors.warning,
      };

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      formatMoney(payout.amountCents, payout.currency),
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w700),
                    ),
                    Text(
                      '${payout.bankName} ${payout.maskedAccountNumber}',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
                    ),
                  ],
                ),
              ),
              AppStatusChip(label: payoutStatusLabel(payout.status), color: _color),
            ],
          ),
          if (payout.status == 'rejected' && payout.rejectReason.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 4),
              child: Text('Lý do: ${payout.rejectReason}', style: const TextStyle(color: AppColors.error, fontSize: 12)),
            ),
        ],
      ),
    );
  }
}
