import 'package:flutter/material.dart';

import 'package:rider/features/booking/domain/models/vehicle_option.dart';

import '../../domain/models/trip_history_entry.dart';
import 'status_chip.dart';

/// Compact Trip History list row: pickup, destination, date, time, fare,
/// status, driver name, vehicle, and rating given.
///
/// Deliberately its own compact widget rather than reusing `PickupCard`/
/// `DestinationCard` (Booking module) here — those are sized for a detail
/// screen, not a dense list of many trips. The full `TripSelection` is
/// still reused as the underlying data (see `TripHistoryEntry.route`), and
/// `PickupCard`/`DestinationCard` ARE reused on `TripDetailPage`.
class TripHistoryTile extends StatelessWidget {
  const TripHistoryTile({super.key, required this.entry, required this.onTap});

  final TripHistoryEntry entry;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    final route = entry.route;

    return InkWell(
      borderRadius: BorderRadius.circular(12),
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.grey.shade200),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(_vehicleIcon(entry.vehicleCategory), size: 18, color: primary),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    '${entry.driver.name} · ${entry.driver.vehicleModel}',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
                  ),
                ),
                StatusChip(status: entry.status),
              ],
            ),
            const SizedBox(height: 10),
            _AddressLine(
              icon: Icons.circle,
              iconColor: primary,
              iconSize: 10,
              text: route.pickupAddress ?? 'Pickup',
            ),
            const SizedBox(height: 4),
            _AddressLine(
              icon: Icons.flag,
              iconColor: Colors.redAccent,
              iconSize: 14,
              text: route.destinationAddress ?? 'Destination',
            ),
            const SizedBox(height: 10),
            Row(
              children: [
                Text(
                  _timeLabel(entry.dateTime),
                  style: TextStyle(fontSize: 12, color: Colors.grey.shade500),
                ),
                const Spacer(),
                if (entry.ratingGiven != null) ...[
                  Icon(Icons.star, size: 14, color: Colors.amber.shade700),
                  const SizedBox(width: 2),
                  Text(
                    entry.ratingGiven!.toStringAsFixed(1),
                    style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
                  ),
                  const SizedBox(width: 10),
                ],
                Text(
                  entry.fare.format(entry.fare.totalCents),
                  style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  static IconData _vehicleIcon(VehicleCategory category) => switch (category) {
        VehicleCategory.car => Icons.directions_car,
        VehicleCategory.motorcycle => Icons.two_wheeler,
        VehicleCategory.van => Icons.airport_shuttle,
      };

  static String _timeLabel(DateTime dt) {
    final hh = dt.hour.toString().padLeft(2, '0');
    final mm = dt.minute.toString().padLeft(2, '0');
    return '$hh:$mm';
  }
}

class _AddressLine extends StatelessWidget {
  const _AddressLine({
    required this.icon,
    required this.iconColor,
    required this.iconSize,
    required this.text,
  });

  final IconData icon;
  final Color iconColor;
  final double iconSize;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        SizedBox(
          width: 16,
          child: Icon(icon, size: iconSize, color: iconColor),
        ),
        const SizedBox(width: 6),
        Expanded(
          child: Text(
            text,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(fontSize: 13, color: Colors.grey.shade700),
          ),
        ),
      ],
    );
  }
}
