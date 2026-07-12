import 'package:flutter/material.dart';

import 'package:driver/core/network/api_client.dart';
import 'package:driver/core/theme/app_spacing.dart';
import 'package:driver/shared/widgets/app_empty_state.dart';

import '../../data/earnings_repository.dart';
import '../../domain/models/earnings_models.dart';
import '../widgets/driver_level_card.dart';
import '../widgets/earnings_breakdown_card.dart';
import '../widgets/earnings_dashboard_card.dart';
import '../widgets/earnings_page_skeleton.dart';
import '../widgets/earnings_quick_actions.dart';
import '../widgets/statistics_grid.dart';
import '../widgets/transaction_history_section.dart';
import '../widgets/wallet_card.dart';

/// Earnings tab — Dashboard, Wallet, Quick Actions, Transaction History,
/// Driver Level, and Statistics, composed from real trip data
/// (`GET /api/v1/driver/trips`, the only earnings-adjacent endpoint that
/// exists) plus honestly-labeled placeholders everywhere the backend has
/// no data source yet (wallet balance, commission breakdown, driver rank,
/// acceptance/completion rate, online hours, distance). See
/// `EarningsRepository` and each widget's doc comment for exactly which
/// numbers are real vs. placeholder.
class EarningsPage extends StatefulWidget {
  const EarningsPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<EarningsPage> createState() => _EarningsPageState();
}

class _EarningsPageState extends State<EarningsPage> {
  late final EarningsRepository _repo;
  EarningsPeriod _period = EarningsPeriod.day;
  late Future<EarningsSummary> _summaryFuture;
  late Future<(int, int)> _allTimeFuture;
  final _transactionsKey = GlobalKey();

  @override
  void initState() {
    super.initState();
    _repo = EarningsRepository(widget.apiClient);
    _summaryFuture = _repo.fetchSummary(_period);
    _allTimeFuture = _repo.fetchAllTimeTripCounts();
  }

  void _onPeriodChanged(EarningsPeriod period) {
    setState(() {
      _period = period;
      _summaryFuture = _repo.fetchSummary(period);
    });
  }

  void _scrollToTransactions() {
    final ctx = _transactionsKey.currentContext;
    if (ctx != null) {
      Scrollable.ensureVisible(ctx, duration: const Duration(milliseconds: 400), curve: Curves.easeOut);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Thu nhập')),
      body: FutureBuilder<EarningsSummary>(
        future: _summaryFuture,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return const EarningsPageSkeleton();
          }
          if (snap.hasError) {
            return AppEmptyState.error(
              subtitle: snap.error is ApiException && (snap.error as ApiException).statusCode == 0
                  ? (snap.error as ApiException).message
                  : 'Không thể tải dữ liệu thu nhập.',
              onAction: () => setState(() => _summaryFuture = _repo.fetchSummary(_period)),
              mascotAsset: 'mascot_no_connection.png',
            );
          }

          final summary = snap.data!;
          return RefreshIndicator(
            onRefresh: () async {
              setState(() {
                _summaryFuture = _repo.fetchSummary(_period);
                _allTimeFuture = _repo.fetchAllTimeTripCounts();
              });
              await Future.wait([_summaryFuture, _allTimeFuture]);
            },
            child: ListView(
              padding: const EdgeInsets.all(AppSpacing.lg),
              children: [
                EarningsDashboardCard(
                  summary: summary,
                  selectedPeriod: _period,
                  onPeriodChanged: _onPeriodChanged,
                ),
                const SizedBox(height: AppSpacing.lg),
                EarningsBreakdownCard(summary: summary),
                const SizedBox(height: AppSpacing.lg),
                const WalletCard(),
                const SizedBox(height: AppSpacing.lg),
                EarningsQuickActions(onViewHistory: _scrollToTransactions),
                const SizedBox(height: AppSpacing.xxl),
                const DriverLevelCard(),
                const SizedBox(height: AppSpacing.xxl),
                Text('Thống kê', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: AppSpacing.md),
                FutureBuilder<(int, int)>(
                  future: _allTimeFuture,
                  builder: (context, statsSnap) {
                    final counts = statsSnap.data ?? (summary.completedCount, summary.cancelledCount);
                    return StatisticsGrid(completedTrips: counts.$1, cancelledTrips: counts.$2);
                  },
                ),
                const SizedBox(height: AppSpacing.xxl),
                KeyedSubtree(
                  key: _transactionsKey,
                  child: TransactionHistorySection(transactions: summary.transactions),
                ),
              ],
            ),
          );
        },
      ),
    );
  }
}
