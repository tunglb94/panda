import 'package:flutter/material.dart';

import '../../domain/models/trip_offer.dart';
import 'route_stat_tile.dart';
import 'trip_address_row.dart';

/// Final screen once dispatch confirms the trip: pickup, destination,
/// estimated fare, Pickup ETA/Distance/status, and a "Start Navigation" CTA.
/// [onNavigate] drives the caller's `TripOfferState` transition to
/// `navigatingToPickup` (Phase D-05) — this widget itself stays state-free,
/// it only reports the tap.
class TripAssignedCard extends StatelessWidget {
  const TripAssignedCard({super.key, required this.offer, required this.onNavigate});

  final TripOffer offer;
  final VoidCallback onNavigate;

  /// Mock ETA for the pickup leg at an assumed ~18 km/h approach speed —
  /// mirrors `RouteProgressModel.mock`'s baseline so the Assigned screen's
  /// static ETA and the Navigation screen's live one agree at the start.
  double get _pickupEtaMinutes => (offer.distanceToPickupKm / 18) * 60;

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
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              Icon(Icons.check_circle, color: primary),
              const SizedBox(width: 8),
              const Text(
                'Trip Assigned',
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 17),
              ),
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
          const SizedBox(height: 16),
          const Divider(height: 1),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: RouteStatTile(
                  icon: Icons.timer_outlined,
                  label: 'Pickup ETA',
                  value: '${_pickupEtaMinutes.round()} mins',
                ),
              ),
              Expanded(
                child: RouteStatTile(
                  icon: Icons.social_distance,
                  label: 'Distance to Pickup',
                  value: '${offer.distanceToPickupKm.toStringAsFixed(1)} km',
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Icon(Icons.info_outline, size: 14, color: Colors.grey.shade500),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  'Current status: Ready to navigate',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          const Divider(height: 1),
          const SizedBox(height: 16),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Estimated fare', style: TextStyle(fontSize: 13, color: Colors.grey.shade600)),
              Text(
                offer.formattedFare,
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18, color: primary),
              ),
            ],
          ),
          const SizedBox(height: 20),
          FilledButton.icon(
            onPressed: onNavigate,
            icon: const Icon(Icons.navigation_outlined),
            label: const Text('Start Navigation'),
          ),
        ],
      ),
    );
  }
}
