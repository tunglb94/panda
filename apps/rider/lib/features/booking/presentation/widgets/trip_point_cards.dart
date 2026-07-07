import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

/// Shared visual for a single trip endpoint (pickup or destination).
/// Not exported directly — use [PickupCard] or [DestinationCard].
class _TripPointCard extends StatelessWidget {
  const _TripPointCard({
    required this.icon,
    required this.iconColor,
    required this.label,
    this.address,
    this.coordinate,
  });

  final IconData icon;
  final Color iconColor;
  final String label;
  final String? address;
  final LatLng? coordinate;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        children: [
          Icon(icon, size: 20, color: iconColor),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: textTheme.labelSmall
                      ?.copyWith(color: Colors.grey.shade500),
                ),
                const SizedBox(height: 2),
                Text(
                  address ?? _formatCoordinate(coordinate),
                  style: textTheme.bodyMedium
                      ?.copyWith(fontWeight: FontWeight.w600),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  static String _formatCoordinate(LatLng? c) {
    if (c == null) return 'Not set';
    return '${c.latitude.toStringAsFixed(5)}, ${c.longitude.toStringAsFixed(5)}';
  }
}

/// Displays the confirmed pickup point of a `TripSelection`.
class PickupCard extends StatelessWidget {
  const PickupCard({super.key, this.address, this.coordinate});

  final String? address;
  final LatLng? coordinate;

  @override
  Widget build(BuildContext context) {
    return _TripPointCard(
      icon: Icons.my_location,
      iconColor: Theme.of(context).colorScheme.primary,
      label: 'Pickup',
      address: address,
      coordinate: coordinate,
    );
  }
}

/// Displays the confirmed destination point of a `TripSelection`.
class DestinationCard extends StatelessWidget {
  const DestinationCard({super.key, this.address, this.coordinate});

  final String? address;
  final LatLng? coordinate;

  @override
  Widget build(BuildContext context) {
    return _TripPointCard(
      icon: Icons.flag,
      iconColor: Colors.redAccent,
      label: 'Destination',
      address: address,
      coordinate: coordinate,
    );
  }
}
