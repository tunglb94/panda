import 'package:flutter/material.dart';

/// A small icon + value + label tile. Shared by `TripAssignedCard` (Pickup
/// ETA / Distance to Pickup) and `DriverNavigationCard` (Distance Remaining
/// / ETA) so the two screens don't duplicate this layout.
class RouteStatTile extends StatelessWidget {
  const RouteStatTile({super.key, required this.icon, required this.label, required this.value});

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Icon(icon, size: 16, color: Colors.grey.shade600),
        const SizedBox(width: 6),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                value,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 13),
              ),
              Text(
                label,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(fontSize: 11, color: Colors.grey.shade500),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
