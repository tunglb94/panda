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
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            blurRadius: 10,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            width: 34,
            height: 34,
            decoration: BoxDecoration(
              color: iconColor.withValues(alpha: 0.12),
              shape: BoxShape.circle,
            ),
            child: Icon(icon, size: 18, color: iconColor),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: textTheme.labelSmall?.copyWith(
                    color: Colors.grey.shade500,
                    fontWeight: FontWeight.w600,
                    letterSpacing: 0.2,
                  ),
                ),
                const SizedBox(height: 3),
                Text(
                  address ?? _formatCoordinate(coordinate),
                  style: textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w700,
                    height: 1.2,
                  ),
                  maxLines: 2,
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
    if (c == null) return 'Chưa đặt';
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
      label: 'Điểm đón',
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
      label: 'Điểm đến',
      address: address,
      coordinate: coordinate,
    );
  }
}

/// Small dashed vertical connector meant to sit between a [PickupCard] and a
/// [DestinationCard] (replacing a plain `SizedBox`), so the two endpoints
/// read as a single route rather than two unrelated boxes.
class RouteConnector extends StatelessWidget {
  const RouteConnector({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 30, top: 2, bottom: 2),
      child: Column(
        children: List.generate(
          3,
          (_) => Padding(
            padding: const EdgeInsets.symmetric(vertical: 1.5),
            child: Container(
              width: 2,
              height: 3,
              decoration: BoxDecoration(
                color: Colors.grey.shade300,
                borderRadius: BorderRadius.circular(1),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
