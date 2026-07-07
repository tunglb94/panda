import 'package:flutter/material.dart';

import '../../domain/models/driver_home_summary.dart';

/// Avatar (initials), name, rating, and vehicle info (or an explicit
/// "no vehicle" message — the Home dashboard's Empty-state condition).
class DriverSummaryHeader extends StatelessWidget {
  const DriverSummaryHeader({super.key, required this.summary});

  final DriverHomeSummary summary;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Row(
      children: [
        CircleAvatar(
          radius: 32,
          backgroundColor: primary.withValues(alpha: 0.12),
          child: Text(
            summary.avatarInitial,
            style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold, color: primary),
          ),
        ),
        const SizedBox(width: 14),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                summary.driverName,
                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 17),
              ),
              const SizedBox(height: 3),
              Row(
                children: [
                  Icon(Icons.star, size: 15, color: Colors.amber.shade700),
                  const SizedBox(width: 4),
                  Text(
                    summary.rating.toStringAsFixed(1),
                    style: TextStyle(fontSize: 13, color: Colors.grey.shade700),
                  ),
                ],
              ),
              const SizedBox(height: 3),
              Row(
                children: [
                  Icon(
                    Icons.directions_car,
                    size: 14,
                    color: summary.hasVehicle ? Colors.grey.shade600 : Colors.orange.shade700,
                  ),
                  const SizedBox(width: 4),
                  Expanded(
                    child: Text(
                      summary.hasVehicle
                          ? '${summary.vehicleModel} · ${summary.plateNumber}'
                          : 'No vehicle assigned yet',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        fontSize: 13,
                        color: summary.hasVehicle ? Colors.grey.shade600 : Colors.orange.shade700,
                        fontWeight: summary.hasVehicle ? FontWeight.normal : FontWeight.w600,
                      ),
                    ),
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
