import 'package:flutter/material.dart';

import '../../domain/models/trip_offer.dart';
import 'rider_info_card.dart';
import 'trip_address_row.dart';

/// Rider info + pickup/destination + distance-to-pickup + surge indicator
/// (only shown when the offer actually has surge pricing).
class TripOfferCard extends StatelessWidget {
  const TripOfferCard({super.key, required this.offer});

  final TripOffer offer;

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
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(child: RiderInfoCard(rider: offer.rider)),
              if (offer.hasSurge) _SurgeBadge(multiplier: offer.surgeMultiplier!),
            ],
          ),
          const SizedBox(height: 16),
          const Divider(height: 1),
          const SizedBox(height: 16),
          TripAddressRow(
            icon: Icons.circle,
            iconSize: 10,
            iconColor: primary,
            label: 'Pickup',
            address: offer.pickupAddress,
          ),
          const SizedBox(height: 10),
          TripAddressRow(
            icon: Icons.flag,
            iconSize: 14,
            iconColor: Colors.redAccent,
            label: 'Destination',
            address: offer.destinationAddress,
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Icon(Icons.near_me_outlined, size: 15, color: Colors.grey.shade600),
              const SizedBox(width: 6),
              Text(
                '${offer.distanceToPickupKm.toStringAsFixed(1)} km to pickup',
                style: TextStyle(fontSize: 13, color: Colors.grey.shade600),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _SurgeBadge extends StatelessWidget {
  const _SurgeBadge({required this.multiplier});

  final double multiplier;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.red.shade50,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.red.shade200),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.bolt, size: 13, color: Colors.red.shade700),
          const SizedBox(width: 3),
          Text(
            '${multiplier.toStringAsFixed(1)}x Surge',
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w700,
              color: Colors.red.shade700,
            ),
          ),
        ],
      ),
    );
  }
}
