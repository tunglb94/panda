import '../../../core/network/api_client.dart';
import '../domain/models/earnings_models.dart';

/// Aggregates [EarningsSummary] client-side from `GET /api/v1/driver/trips`
/// — the exact same endpoint `DriverTripHistoryPage` already calls. No new
/// endpoint, no new backend field. Anything the backend doesn't expose
/// (wallet balance, commission breakdown, bonuses, online hours,
/// acceptance/completion rate...) is deliberately left out of this
/// repository rather than estimated or invented.
class EarningsRepository {
  const EarningsRepository(this._client);

  final ApiClient _client;

  Future<List<EarningsTransaction>> _fetchAllTrips() async {
    final body = await _client.get('/api/v1/driver/trips');
    final raw = (body['trips'] as List<dynamic>?) ?? [];
    return raw.map((e) => _fromJson(e as Map<String, dynamic>)).toList()
      ..sort((a, b) => b.createdAt.compareTo(a.createdAt));
  }

  EarningsTransaction _fromJson(Map<String, dynamic> j) {
    DateTime dt;
    try {
      dt = DateTime.parse(j['created_at'] as String? ?? '');
    } catch (_) {
      dt = DateTime.fromMillisecondsSinceEpoch(0);
    }
    return EarningsTransaction(
      tripId: j['trip_id'] as String? ?? '',
      status: j['status'] as String? ?? '',
      pickupAddress: j['pickup_address'] as String? ?? '',
      dropoffAddress: j['dropoff_address'] as String? ?? '',
      amountCents: (j['final_fare'] as num?)?.toInt() ?? 0,
      currency: j['currency'] as String? ?? '',
      createdAt: dt.toLocal(),
    );
  }

  /// All-time completed/cancelled trip counts (not period-scoped) — used by
  /// the Statistics grid's one real card ("Chuyến đi"). Everything else in
  /// that grid (acceptance rate, online hours, distance, rating) has no
  /// backend source at all, period-scoped or not.
  Future<(int completed, int cancelled)> fetchAllTimeTripCounts() async {
    final all = await _fetchAllTrips();
    var completed = 0;
    var cancelled = 0;
    for (final t in all) {
      if (t.isEarning) completed++;
      if (t.isCancelled) cancelled++;
    }
    return (completed, cancelled);
  }

  Future<EarningsSummary> fetchSummary(EarningsPeriod period) async {
    final all = await _fetchAllTrips();
    final now = DateTime.now();
    final rangeStart = switch (period) {
      EarningsPeriod.day => DateTime(now.year, now.month, now.day),
      EarningsPeriod.week => now.subtract(Duration(days: now.weekday - 1)),
      EarningsPeriod.month => DateTime(now.year, now.month, 1),
    };
    final inRange = all.where((t) => !t.createdAt.isBefore(rangeStart)).toList();

    var totalCents = 0;
    var currency = '';
    var completed = 0;
    var cancelled = 0;
    for (final t in inRange) {
      if (t.isEarning) {
        totalCents += t.amountCents;
        if (currency.isEmpty && t.currency.isNotEmpty) currency = t.currency;
        completed++;
      } else if (t.isCancelled) {
        cancelled++;
      }
    }

    // Last 7 days, always computed from the full trip list regardless of
    // the selected period tab — the chart is a fixed "trend" view.
    final sevenDaysAgo = DateTime(now.year, now.month, now.day)
        .subtract(const Duration(days: 6));
    final byDay = <DateTime, int>{};
    for (var i = 0; i < 7; i++) {
      byDay[DateTime(sevenDaysAgo.year, sevenDaysAgo.month, sevenDaysAgo.day + i)] = 0;
    }
    for (final t in all) {
      if (!t.isEarning) continue;
      final day = DateTime(t.createdAt.year, t.createdAt.month, t.createdAt.day);
      if (byDay.containsKey(day)) {
        byDay[day] = (byDay[day] ?? 0) + t.amountCents;
      }
    }
    final dailySeries = byDay.entries
        .map((e) => DailyEarningsPoint(day: e.key, amountCents: e.value))
        .toList()
      ..sort((a, b) => a.day.compareTo(b.day));

    return EarningsSummary(
      period: period,
      totalCents: totalCents,
      currency: currency,
      completedCount: completed,
      cancelledCount: cancelled,
      dailySeries: dailySeries,
      transactions: inRange,
    );
  }
}
