import 'package:flutter/material.dart';

/// Date-group header shown above a batch of trips sharing the same day
/// (e.g. "Today", "Yesterday", "Jul 3, 2026").
class TripHistorySectionHeader extends StatelessWidget {
  const TripHistorySectionHeader({super.key, required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8, top: 4),
      child: Text(
        label,
        style: const TextStyle(
          fontSize: 13,
          fontWeight: FontWeight.w700,
          color: Color(0xFF6B7280),
        ),
      ),
    );
  }

  static const _months = [
    'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun',
    'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec',
  ];

  /// Groups trips by calendar day: "Today", "Yesterday", or "MMM d, yyyy".
  /// Shared by `TripHistoryPage` (to build the grouped list) and this
  /// header (to render the same label).
  static String labelFor(DateTime dateTime) {
    final now = DateTime.now();
    final day = DateTime(dateTime.year, dateTime.month, dateTime.day);
    final today = DateTime(now.year, now.month, now.day);
    final diff = today.difference(day).inDays;
    if (diff == 0) return 'Today';
    if (diff == 1) return 'Yesterday';
    return '${_months[dateTime.month - 1]} ${dateTime.day}, ${dateTime.year}';
  }
}
