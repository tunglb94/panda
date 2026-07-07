import 'package:flutter/material.dart';

import '../../domain/models/route_progress_model.dart';
import '../../domain/models/trip_offer.dart';
import 'driver_status_banner.dart';
import 'route_progress_indicator.dart';
import 'route_stat_tile.dart';
import 'trip_address_row.dart';

/// The Navigation screen shown once the driver taps "Start Navigation" on
/// `TripAssignedCard` (Phase D-05): "Driving to Pickup" status, the pickup
/// address, distance/ETA remaining, mock route progress, and Contact
/// Rider/Cancel Trip actions. No map, no GPS — [route] is a mock snapshot.
class DriverNavigationCard extends StatelessWidget {
  const DriverNavigationCard({
    super.key,
    required this.offer,
    required this.route,
    required this.onContactRider,
    required this.onCancelTrip,
  });

  final TripOffer offer;
  final RouteProgressModel route;
  final VoidCallback onContactRider;
  final VoidCallback onCancelTrip;

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
          DriverStatusBanner(
            icon: Icons.navigation,
            title: 'Driving to Pickup',
            subtitle: "Heading to ${offer.rider.name}'s pickup point",
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
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: RouteStatTile(
                  icon: Icons.social_distance,
                  label: 'Distance remaining',
                  value: '${route.remainingDistanceKm.toStringAsFixed(1)} km',
                ),
              ),
              Expanded(
                child: RouteStatTile(
                  icon: Icons.timer_outlined,
                  label: 'ETA',
                  value: '${route.remainingDurationMin.round()} min',
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          RouteProgressIndicator(progress: route.progress, trafficLevel: route.trafficLevel),
          const SizedBox(height: 20),
          Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: onCancelTrip,
                  icon: const Icon(Icons.close),
                  label: const Text('Cancel Trip'),
                  style: OutlinedButton.styleFrom(foregroundColor: Colors.red.shade600),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: FilledButton.icon(
                  onPressed: onContactRider,
                  icon: const Icon(Icons.call_outlined),
                  label: const Text('Contact Rider'),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
