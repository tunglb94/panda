import 'package:flutter/material.dart';

import '../../domain/models/driver_home_summary.dart';

/// Today's completed trips, earnings (mock), and online duration (mock) —
/// three-column stat row, styled like `apps/rider`'s `ProfileStatsRow` /
/// `EtaArrivalCard` (Profile/Trip modules) for a consistent design language.
class HomeStatsRow extends StatelessWidget {
  const HomeStatsRow({super.key, required this.summary});

  final DriverHomeSummary summary;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        children: [
          Expanded(
            child: _Stat(
              icon: Icons.route,
              label: 'Trips today',
              value: '${summary.completedTripsToday}',
            ),
          ),
          Container(width: 1, height: 34, color: Colors.grey.shade300),
          Expanded(
            child: _Stat(
              icon: Icons.payments_outlined,
              label: 'Earnings today',
              value: summary.formattedEarningsToday,
            ),
          ),
          Container(width: 1, height: 34, color: Colors.grey.shade300),
          Expanded(
            child: _Stat(
              icon: Icons.timer_outlined,
              label: 'Online time',
              value: summary.formattedOnlineDuration,
            ),
          ),
        ],
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat({required this.icon, required this.label, required this.value});

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Column(
      children: [
        Icon(icon, size: 18, color: primary),
        const SizedBox(height: 6),
        Text(
          value,
          style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
        ),
        const SizedBox(height: 2),
        Text(
          label,
          textAlign: TextAlign.center,
          style: TextStyle(fontSize: 11, color: Colors.grey.shade500),
        ),
      ],
    );
  }
}
