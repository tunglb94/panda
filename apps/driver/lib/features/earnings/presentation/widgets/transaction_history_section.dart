import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_empty_state.dart';
import '../../../../shared/widgets/app_status_chip.dart';
import '../../domain/models/earnings_models.dart';

enum _StatusFilter { all, completed, cancelled }

/// Real transaction list — every row is an actual trip from
/// `GET /api/v1/driver/trips`. The search box is fully functional (filters
/// the already-fetched list by address text — no new API call, no
/// server-side search), and the status filter chips are real client-side
/// filtering too, not placeholders.
class TransactionHistorySection extends StatefulWidget {
  const TransactionHistorySection({super.key, required this.transactions});

  final List<EarningsTransaction> transactions;

  @override
  State<TransactionHistorySection> createState() => _TransactionHistorySectionState();
}

class _TransactionHistorySectionState extends State<TransactionHistorySection> {
  _StatusFilter _filter = _StatusFilter.all;
  final _searchController = TextEditingController();
  String _query = '';

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  List<EarningsTransaction> get _filtered {
    return widget.transactions.where((t) {
      final matchesFilter = switch (_filter) {
        _StatusFilter.all => true,
        _StatusFilter.completed => t.isEarning,
        _StatusFilter.cancelled => t.isCancelled,
      };
      if (!matchesFilter) return false;
      if (_query.isEmpty) return true;
      final q = _query.toLowerCase();
      return t.pickupAddress.toLowerCase().contains(q) ||
          t.dropoffAddress.toLowerCase().contains(q);
    }).toList();
  }

  @override
  Widget build(BuildContext context) {
    final items = _filtered;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Lịch sử giao dịch', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: AppSpacing.sm),
        TextField(
          controller: _searchController,
          onChanged: (v) => setState(() => _query = v),
          decoration: InputDecoration(
            hintText: 'Tìm theo địa chỉ đón/trả…',
            prefixIcon: const Icon(Icons.search, size: 20),
            suffixIcon: _query.isEmpty
                ? null
                : IconButton(
                    icon: const Icon(Icons.clear, size: 18),
                    tooltip: 'Xóa tìm kiếm',
                    onPressed: () => setState(() {
                      _searchController.clear();
                      _query = '';
                    }),
                  ),
          ),
        ),
        const SizedBox(height: AppSpacing.sm),
        Row(
          children: [
            _FilterChip(
              label: 'Tất cả',
              selected: _filter == _StatusFilter.all,
              onTap: () => setState(() => _filter = _StatusFilter.all),
            ),
            const SizedBox(width: AppSpacing.sm),
            _FilterChip(
              label: 'Hoàn thành',
              selected: _filter == _StatusFilter.completed,
              onTap: () => setState(() => _filter = _StatusFilter.completed),
            ),
            const SizedBox(width: AppSpacing.sm),
            _FilterChip(
              label: 'Đã hủy',
              selected: _filter == _StatusFilter.cancelled,
              onTap: () => setState(() => _filter = _StatusFilter.cancelled),
            ),
          ],
        ),
        const SizedBox(height: AppSpacing.md),
        if (items.isEmpty)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: AppSpacing.xl),
            child: AppEmptyState(
              icon: Icons.receipt_long_outlined,
              title: 'Không có giao dịch',
              subtitle: 'Không tìm thấy giao dịch phù hợp với bộ lọc hiện tại.',
              // Mascot only when there are truly no transactions yet, not
              // when a search/filter narrows an existing list down to zero.
              mascotAsset: widget.transactions.isEmpty ? 'mascot_waiting.png' : null,
            ),
          )
        else
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: items.length,
            separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
            itemBuilder: (context, i) => _TransactionTile(transaction: items[i]),
          ),
      ],
    );
  }
}

class _FilterChip extends StatelessWidget {
  const _FilterChip({required this.label, required this.selected, required this.onTap});

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 180),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 7),
        decoration: BoxDecoration(
          color: selected ? AppColors.primary : AppColors.surfaceAlt,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: selected ? AppColors.primary : AppColors.border),
        ),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 12,
            fontWeight: FontWeight.w600,
            color: selected ? Colors.white : AppColors.textSecondary,
          ),
        ),
      ),
    );
  }
}

class _TransactionTile extends StatelessWidget {
  const _TransactionTile({required this.transaction});

  final EarningsTransaction transaction;

  @override
  Widget build(BuildContext context) {
    final (icon, color, label) = switch (transaction.status) {
      'completed' || 'settled' => (Icons.check_circle, AppColors.primary, 'Hoàn thành'),
      'cancelled' => (Icons.cancel, AppColors.textTertiary, 'Đã hủy'),
      'in_progress' => (Icons.directions_car, AppColors.info, 'Đang thực hiện'),
      'payment_pending' || 'payment_success' => (Icons.schedule, AppColors.warning, 'Chờ thanh toán'),
      _ => (Icons.directions_car, AppColors.textSecondary, 'Đang thực hiện'),
    };

    return AppCard(
      animateIn: false,
      padding: const EdgeInsets.all(AppSpacing.md),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(color: color.withValues(alpha: 0.12), shape: BoxShape.circle),
            child: Icon(icon, color: color, size: AppIconSize.md),
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  transaction.dropoffAddress.isEmpty
                      ? transaction.pickupAddress
                      : transaction.dropoffAddress,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
                ),
                const SizedBox(height: 2),
                Text(_formatDate(transaction.createdAt), style: Theme.of(context).textTheme.bodySmall),
              ],
            ),
          ),
          const SizedBox(width: AppSpacing.sm),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                transaction.amountLabel,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      fontWeight: FontWeight.w700,
                      color: transaction.isEarning ? AppColors.primary : AppColors.textTertiary,
                    ),
              ),
              const SizedBox(height: 2),
              AppStatusChip(label: label, color: color),
            ],
          ),
        ],
      ),
    );
  }

  static String _formatDate(DateTime dt) {
    final now = DateTime.now();
    final hh = dt.hour.toString().padLeft(2, '0');
    final mm = dt.minute.toString().padLeft(2, '0');
    if (dt.year == now.year && dt.month == now.month && dt.day == now.day) {
      return 'Hôm nay, $hh:$mm';
    }
    return '${dt.day}/${dt.month}/${dt.year}, $hh:$mm';
  }
}
