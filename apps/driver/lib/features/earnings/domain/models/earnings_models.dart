import '../../../../shared/utils/currency_format.dart';

/// A single trip mapped into an earnings "transaction" row. This is a
/// straight re-projection of `GET /api/v1/driver/trips` (the same endpoint
/// `DriverTripHistoryPage` already uses) — no new endpoint, no invented
/// fields. `final_fare`/`currency` are the only money fields the backend
/// exposes today; there is no gross-fare/commission/tax breakdown per trip
/// (see `docs/driver/DRIVER_APP_SPEC.md` §11.4 — that needs new backend
/// work this task explicitly does not do).
class EarningsTransaction {
  const EarningsTransaction({
    required this.tripId,
    required this.status,
    required this.pickupAddress,
    required this.dropoffAddress,
    required this.amountCents,
    required this.currency,
    required this.createdAt,
  });

  final String tripId;
  final String status;
  final String pickupAddress;
  final String dropoffAddress;
  final int amountCents;
  final String currency;
  final DateTime createdAt;

  /// Only settled/completed trips represent real, countable earnings.
  bool get isEarning => status == 'completed' || status == 'settled';
  bool get isCancelled => status == 'cancelled';

  String get amountLabel {
    if (amountCents <= 0 || currency.isEmpty) return '—';
    return formatMoney(amountCents, currency);
  }
}

/// One day's worth of real earnings, used to draw the mini bar/line chart.
/// Always derived from actual [EarningsTransaction]s — never synthesized.
class DailyEarningsPoint {
  const DailyEarningsPoint({required this.day, required this.amountCents});

  final DateTime day;
  final int amountCents;
}

enum EarningsPeriod { day, week, month }

/// Aggregated totals for a period, computed client-side from the full trip
/// list — the backend has no `/earnings` endpoint (and this task adds none).
class EarningsSummary {
  const EarningsSummary({
    required this.period,
    required this.totalCents,
    required this.currency,
    required this.completedCount,
    required this.cancelledCount,
    required this.dailySeries,
    required this.transactions,
  });

  final EarningsPeriod period;
  final int totalCents;
  final String currency;
  final int completedCount;
  final int cancelledCount;
  final List<DailyEarningsPoint> dailySeries;
  final List<EarningsTransaction> transactions;

  String get totalLabel {
    if (totalCents <= 0 || currency.isEmpty) return '—';
    return formatMoney(totalCents, currency);
  }

  bool get isEmpty => transactions.isEmpty;
}
