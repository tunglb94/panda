import 'package:flutter/material.dart';

import '../../domain/models/trip_history_filters.dart';

/// Search box + status/date filter chips for `TripHistoryPage`. Filtering
/// is entirely client-side over the already-fetched mock list — there is no
/// search or filter backend.
class HistoryFilterBar extends StatelessWidget {
  const HistoryFilterBar({
    super.key,
    required this.query,
    required this.onQueryChanged,
    required this.statusFilter,
    required this.onStatusChanged,
    required this.dateFilter,
    required this.onDateChanged,
  });

  final String query;
  final ValueChanged<String> onQueryChanged;
  final TripHistoryStatusFilter statusFilter;
  final ValueChanged<TripHistoryStatusFilter> onStatusChanged;
  final TripHistoryDateFilter dateFilter;
  final ValueChanged<TripHistoryDateFilter> onDateChanged;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        TextField(
          onChanged: onQueryChanged,
          decoration: InputDecoration(
            hintText: 'Search by address or driver',
            isDense: true,
            prefixIcon: const Icon(Icons.search),
            suffixIcon: query.isEmpty
                ? null
                : IconButton(
                    icon: const Icon(Icons.close),
                    onPressed: () => onQueryChanged(''),
                  ),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
          ),
        ),
        const SizedBox(height: 10),
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: Row(
            children: [
              for (final filter in TripHistoryStatusFilter.values) ...[
                ChoiceChip(
                  label: Text(filter.label),
                  selected: statusFilter == filter,
                  onSelected: (_) => onStatusChanged(filter),
                ),
                const SizedBox(width: 8),
              ],
            ],
          ),
        ),
        const SizedBox(height: 8),
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: Row(
            children: [
              for (final filter in TripHistoryDateFilter.values) ...[
                ChoiceChip(
                  label: Text(filter.label),
                  selected: dateFilter == filter,
                  onSelected: (_) => onDateChanged(filter),
                ),
                const SizedBox(width: 8),
              ],
            ],
          ),
        ),
      ],
    );
  }
}
