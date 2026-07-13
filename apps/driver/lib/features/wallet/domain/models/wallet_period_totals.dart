import 'wallet_transaction.dart';

/// Client-side period aggregation from the driver's own transaction list —
/// same approach `EarningsRepository` already uses for the trip-derived
/// dashboard (no new backend endpoint for "today/week/month" needed; Phần 9
/// full-precision Driver Statement is available separately via
/// `WalletRepository.fetchStatement` for a future export feature).
const _incomeTypes = {'ride_income', 'delivery_income', 'bonus'};

int _sumIncomeSince(List<WalletTransaction> txs, DateTime since) {
  var total = 0;
  for (final t in txs) {
    if (t.createdAt == null) continue;
    if (!t.isCredit || !_incomeTypes.contains(t.type)) continue;
    if (t.createdAt!.isBefore(since)) continue;
    total += t.amountCents;
  }
  return total;
}

class WalletPeriodTotals {
  const WalletPeriodTotals({required this.todayCents, required this.weekCents, required this.monthCents});

  final int todayCents;
  final int weekCents;
  final int monthCents;

  factory WalletPeriodTotals.fromTransactions(List<WalletTransaction> txs, {DateTime? now}) {
    final today = now ?? DateTime.now();
    final startOfToday = DateTime(today.year, today.month, today.day);
    final startOfWeek = startOfToday.subtract(Duration(days: 6));
    final startOfMonth = startOfToday.subtract(const Duration(days: 29));
    return WalletPeriodTotals(
      todayCents: _sumIncomeSince(txs, startOfToday),
      weekCents: _sumIncomeSince(txs, startOfWeek),
      monthCents: _sumIncomeSince(txs, startOfMonth),
    );
  }

  /// Last 7 days' daily income totals, oldest first — feeds the Phần 7 line chart.
  static List<int> dailySeries(List<WalletTransaction> txs, {DateTime? now}) {
    final today = now ?? DateTime.now();
    final startOfToday = DateTime(today.year, today.month, today.day);
    return List.generate(7, (i) {
      final day = startOfToday.subtract(Duration(days: 6 - i));
      final nextDay = day.add(const Duration(days: 1));
      var total = 0;
      for (final t in txs) {
        if (t.createdAt == null) continue;
        if (!t.isCredit || !_incomeTypes.contains(t.type)) continue;
        if (t.createdAt!.isBefore(day) || !t.createdAt!.isBefore(nextDay)) continue;
        total += t.amountCents;
      }
      return total;
    });
  }
}
