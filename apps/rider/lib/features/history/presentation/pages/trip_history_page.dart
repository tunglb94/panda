import 'package:flutter/material.dart';

import 'package:rider/features/profile/presentation/widgets/async_state_view.dart';

import '../../domain/models/mock_trip_history_repository.dart';
import '../../domain/models/trip_history_entry.dart';
import '../../domain/models/trip_history_filters.dart';
import '../../domain/models/trip_history_status.dart';
import '../widgets/history_filter_bar.dart';
import '../widgets/trip_history_section_header.dart';
import '../widgets/trip_history_tile.dart';
import 'trip_detail_page.dart';

/// Trip History screen: search + status/date filters over a mock trip list,
/// grouped by day. Reuses `AsyncStateView` (Profile module, R-03) for the
/// Loading/Success/Empty/Error states.
class TripHistoryPage extends StatefulWidget {
  const TripHistoryPage({super.key});

  @override
  State<TripHistoryPage> createState() => _TripHistoryPageState();
}

class _TripHistoryPageState extends State<TripHistoryPage> {
  static const _repository = MockTripHistoryRepository();

  TripHistoryDemoMode _mode = TripHistoryDemoMode.normal;
  late Future<List<TripHistoryEntry>> _future;

  String _query = '';
  TripHistoryStatusFilter _statusFilter = TripHistoryStatusFilter.all;
  TripHistoryDateFilter _dateFilter = TripHistoryDateFilter.all;

  @override
  void initState() {
    super.initState();
    _load();
  }

  void _load() {
    setState(() {
      _future = _repository.fetchHistory(mode: _mode);
    });
  }

  List<TripHistoryEntry> _applyFilters(List<TripHistoryEntry> all) {
    return all.where((entry) {
      final statusOk = switch (_statusFilter) {
        TripHistoryStatusFilter.all => true,
        TripHistoryStatusFilter.completed => entry.status == TripHistoryStatus.completed,
        TripHistoryStatusFilter.cancelled => entry.status == TripHistoryStatus.cancelled,
      };
      final dateOk = _dateFilter.matches(entry.dateTime);
      final queryOk = entry.matchesQuery(_query);
      return statusOk && dateOk && queryOk;
    }).toList()
      ..sort((a, b) => b.dateTime.compareTo(a.dateTime));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Trip History'),
        actions: [
          PopupMenuButton<TripHistoryDemoMode>(
            tooltip: 'Preview state (dev)',
            icon: const Icon(Icons.tune),
            onSelected: (mode) {
              _mode = mode;
              _load();
            },
            itemBuilder: (context) => const [
              PopupMenuItem(value: TripHistoryDemoMode.normal, child: Text('Normal')),
              PopupMenuItem(value: TripHistoryDemoMode.empty, child: Text('Empty (dev)')),
              PopupMenuItem(value: TripHistoryDemoMode.error, child: Text('Error (dev)')),
            ],
          ),
        ],
      ),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: Column(
              children: [
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
                  child: HistoryFilterBar(
                    query: _query,
                    onQueryChanged: (q) => setState(() => _query = q),
                    statusFilter: _statusFilter,
                    onStatusChanged: (f) => setState(() => _statusFilter = f),
                    dateFilter: _dateFilter,
                    onDateChanged: (f) => setState(() => _dateFilter = f),
                  ),
                ),
                Expanded(
                  child: AsyncStateView<List<TripHistoryEntry>>(
                    future: _future,
                    isEmpty: (items) => items.isEmpty,
                    emptyBuilder: (context) => Padding(
                      padding: const EdgeInsets.symmetric(vertical: 48, horizontal: 24),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.receipt_long_outlined, size: 48, color: Colors.grey.shade400),
                          const SizedBox(height: 12),
                          const Text('No trips yet', style: TextStyle(fontWeight: FontWeight.w600)),
                          const SizedBox(height: 4),
                          Text(
                            'Your completed and cancelled rides will show up here.',
                            textAlign: TextAlign.center,
                            style: TextStyle(color: Colors.grey.shade500),
                          ),
                        ],
                      ),
                    ),
                    errorBuilder: (context, error) => Padding(
                      padding: const EdgeInsets.symmetric(vertical: 48, horizontal: 24),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.error_outline, size: 48, color: Colors.red.shade400),
                          const SizedBox(height: 12),
                          const Text("Couldn't load trip history",
                              style: TextStyle(fontWeight: FontWeight.w600)),
                          const SizedBox(height: 12),
                          OutlinedButton(onPressed: _load, child: const Text('Retry')),
                        ],
                      ),
                    ),
                    successBuilder: (context, all) {
                      final filtered = _applyFilters(all);
                      if (filtered.isEmpty) {
                        return Padding(
                          padding: const EdgeInsets.symmetric(vertical: 48, horizontal: 24),
                          child: Column(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(Icons.filter_alt_off_outlined,
                                  size: 40, color: Colors.grey.shade400),
                              const SizedBox(height: 12),
                              const Text('No trips match these filters',
                                  style: TextStyle(fontWeight: FontWeight.w600)),
                              const SizedBox(height: 12),
                              TextButton(
                                onPressed: () => setState(() {
                                  _query = '';
                                  _statusFilter = TripHistoryStatusFilter.all;
                                  _dateFilter = TripHistoryDateFilter.all;
                                }),
                                child: const Text('Clear filters'),
                              ),
                            ],
                          ),
                        );
                      }

                      final groups = <String, List<TripHistoryEntry>>{};
                      for (final entry in filtered) {
                        final label = TripHistorySectionHeader.labelFor(entry.dateTime);
                        groups.putIfAbsent(label, () => []).add(entry);
                      }

                      return ListView(
                        padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
                        children: [
                          for (final group in groups.entries) ...[
                            TripHistorySectionHeader(label: group.key),
                            for (final entry in group.value) ...[
                              TripHistoryTile(
                                entry: entry,
                                onTap: () => Navigator.of(context).push(
                                  MaterialPageRoute(
                                    builder: (_) => TripDetailPage(entry: entry),
                                  ),
                                ),
                              ),
                              const SizedBox(height: 10),
                            ],
                          ],
                        ],
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
