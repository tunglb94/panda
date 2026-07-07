import 'package:flutter/material.dart';

import '../../domain/models/rider_info.dart';

/// Rider avatar (initials), name, and rating.
class RiderInfoCard extends StatelessWidget {
  const RiderInfoCard({super.key, required this.rider});

  final RiderInfo rider;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Row(
      children: [
        CircleAvatar(
          radius: 24,
          backgroundColor: primary.withValues(alpha: 0.12),
          child: Text(
            rider.avatarInitial,
            style: TextStyle(fontSize: 17, fontWeight: FontWeight.bold, color: primary),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(rider.name, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 15)),
              const SizedBox(height: 2),
              Row(
                children: [
                  Icon(Icons.star, size: 14, color: Colors.amber.shade700),
                  const SizedBox(width: 4),
                  Text(
                    rider.rating.toStringAsFixed(1),
                    style: TextStyle(fontSize: 13, color: Colors.grey.shade700),
                  ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }
}
