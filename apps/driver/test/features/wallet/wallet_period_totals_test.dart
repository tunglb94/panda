// Unit tests for the Driver Finance module's client-side period
// aggregation (Phần 7 — Hôm nay/Tuần/Tháng cards + 7-day chart), which is
// derived entirely from the driver's own transaction list, never computed
// by re-deriving ledger math client-side.
import 'package:flutter_test/flutter_test.dart';

import 'package:driver/features/wallet/domain/models/wallet_period_totals.dart';
import 'package:driver/features/wallet/domain/models/wallet_transaction.dart';

WalletTransaction _income(DateTime createdAt, int amountCents, {String type = 'ride_income'}) => WalletTransaction(
      type: type,
      direction: 'credit',
      amountCents: amountCents,
      currency: 'VND',
      description: '',
      paymentMethod: 'wallet',
      createdAt: createdAt,
    );

void main() {
  final now = DateTime(2026, 7, 10, 15, 0);

  test('fromTransactions sums only today into todayCents', () {
    final txs = [
      _income(DateTime(2026, 7, 10, 8, 0), 100_000), // today
      _income(DateTime(2026, 7, 9, 20, 0), 50_000), // yesterday
    ];
    final totals = WalletPeriodTotals.fromTransactions(txs, now: now);
    expect(totals.todayCents, 100_000);
    expect(totals.weekCents, 150_000);
  });

  test('fromTransactions excludes debit and non-income types', () {
    final txs = [
      _income(now, 100_000),
      WalletTransaction(
        type: 'ride_income',
        direction: 'debit', // e.g. a reversal — must not count as income
        amountCents: 50_000,
        currency: 'VND',
        description: '',
        paymentMethod: 'wallet',
        createdAt: now,
      ),
      WalletTransaction(
        type: 'withdrawal',
        direction: 'debit',
        amountCents: 30_000,
        currency: 'VND',
        description: '',
        paymentMethod: '',
        createdAt: now,
      ),
    ];
    final totals = WalletPeriodTotals.fromTransactions(txs, now: now);
    expect(totals.todayCents, 100_000);
  });

  test('fromTransactions excludes transactions older than the month window', () {
    final txs = [
      _income(now.subtract(const Duration(days: 40)), 999_000),
    ];
    final totals = WalletPeriodTotals.fromTransactions(txs, now: now);
    expect(totals.monthCents, 0);
  });

  test('dailySeries buckets income by day, oldest first, 7 entries', () {
    final txs = [
      _income(now, 10_000), // today
      _income(now.subtract(const Duration(days: 6)), 20_000), // 6 days ago (first bucket)
    ];
    final series = WalletPeriodTotals.dailySeries(txs, now: now);
    expect(series.length, 7);
    expect(series.first, 20_000);
    expect(series.last, 10_000);
  });

  test('dailySeries is all zeros for an empty transaction list', () {
    final series = WalletPeriodTotals.dailySeries(const [], now: now);
    expect(series, List.filled(7, 0));
  });
}
