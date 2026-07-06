import 'package:flutter/material.dart';

import 'package:rider/features/booking/domain/models/vehicle_option.dart';
import 'package:rider/features/trip/domain/models/mock_driver.dart';

/// "Vehicle information" block on the Trip Detail screen — kept separate
/// from `DriverInfoCard` (which already shows model/plate inline) so
/// vehicle details have their own clearly labelled section, as requested.
class VehicleInfoCard extends StatelessWidget {
  const VehicleInfoCard({super.key, required this.category, required this.driver});

  final VehicleCategory category;
  final MockDriver driver;

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
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: primary.withValues(alpha: 0.1),
              shape: BoxShape.circle,
            ),
            child: Icon(_icon, color: primary),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(driver.vehicleModel,
                    style: const TextStyle(fontWeight: FontWeight.w600)),
                const SizedBox(height: 2),
                Text(
                  '$_label · ${driver.plateNumber}',
                  style: TextStyle(fontSize: 13, color: Colors.grey.shade600),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  IconData get _icon => switch (category) {
        VehicleCategory.car => Icons.directions_car,
        VehicleCategory.motorcycle => Icons.two_wheeler,
        VehicleCategory.van => Icons.airport_shuttle,
      };

  String get _label => switch (category) {
        VehicleCategory.car => 'Car',
        VehicleCategory.motorcycle => 'Motorcycle',
        VehicleCategory.van => 'Van',
      };
}
