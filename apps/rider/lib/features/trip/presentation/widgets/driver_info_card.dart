import 'package:flutter/material.dart';

import '../../domain/models/driver_profile.dart';

class DriverInfoCard extends StatelessWidget {
  const DriverInfoCard({super.key, required this.driver});

  final DriverProfile driver;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        children: [
          CircleAvatar(
            radius: 26,
            backgroundColor: primary.withValues(alpha: 0.12),
            child: Icon(Icons.directions_car, color: primary, size: 24),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  driver.vehicleDisplay,
                  style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 15),
                ),
                if (driver.vehicleColor.isNotEmpty) ...[
                  const SizedBox(height: 3),
                  Text(
                    driver.vehicleColor,
                    style: TextStyle(fontSize: 13, color: Colors.grey.shade600),
                  ),
                ],
                const SizedBox(height: 3),
                Text(
                  driver.plateNumber,
                  style: TextStyle(fontSize: 13, color: Colors.grey.shade700),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
