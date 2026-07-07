import 'package:flutter/material.dart';

/// A single labelled address line (Pickup or Destination), with a
/// small colored marker icon. Shared by `TripOfferCard` (D-03) and
/// `TripAssignedCard` (D-04) so the two don't duplicate this layout.
class TripAddressRow extends StatelessWidget {
  const TripAddressRow({
    super.key,
    required this.icon,
    required this.iconSize,
    required this.iconColor,
    required this.label,
    required this.address,
  });

  final IconData icon;
  final double iconSize;
  final Color iconColor;
  final String label;
  final String address;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(width: 18, child: Icon(icon, size: iconSize, color: iconColor)),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label, style: TextStyle(fontSize: 11, color: Colors.grey.shade500)),
              Text(
                address,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(fontWeight: FontWeight.w500),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
