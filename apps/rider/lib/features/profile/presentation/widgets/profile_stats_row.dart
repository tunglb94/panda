import 'package:flutter/material.dart';

import '../../domain/models/rider_profile.dart';

/// Rating and total completed trips, side by side.
class ProfileStatsRow extends StatelessWidget {
  const ProfileStatsRow({super.key, required this.profile});

  final RiderProfile profile;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        children: [
          Expanded(
            child: _StatColumn(
              icon: Icons.star,
              iconColor: Colors.amber.shade700,
              label: 'Rating',
              value: profile.rating.toStringAsFixed(1),
            ),
          ),
          Container(width: 1, height: 36, color: Colors.grey.shade300),
          Expanded(
            child: _StatColumn(
              icon: Icons.route,
              iconColor: Theme.of(context).colorScheme.primary,
              label: 'Completed Trips',
              value: '${profile.totalCompletedTrips}',
            ),
          ),
        ],
      ),
    );
  }
}

class _StatColumn extends StatelessWidget {
  const _StatColumn({
    required this.icon,
    required this.iconColor,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final Color iconColor;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, size: 16, color: iconColor),
            const SizedBox(width: 4),
            Text(value, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
          ],
        ),
        const SizedBox(height: 2),
        Text(label, style: TextStyle(fontSize: 11, color: Colors.grey.shade500)),
      ],
    );
  }
}
