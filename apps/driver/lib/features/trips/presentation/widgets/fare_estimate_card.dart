import 'package:flutter/material.dart';

import '../../domain/models/trip_offer.dart';
import 'route_stat_tile.dart';

/// Estimated fare (headline) + trip distance and duration.
class FareEstimateCard extends StatelessWidget {
  const FareEstimateCard({super.key, required this.offer});

  final TripOffer offer;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Estimated fare',
                style: TextStyle(fontSize: 13, color: Colors.grey.shade600),
              ),
              Text(
                offer.formattedFare,
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 20, color: primary),
              ),
            ],
          ),
          const SizedBox(height: 12),
          const Divider(height: 1),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: RouteStatTile(
                  icon: Icons.route,
                  label: 'Trip distance',
                  value: '${offer.estimatedTripDistanceKm.toStringAsFixed(1)} km',
                ),
              ),
              Expanded(
                child: RouteStatTile(
                  icon: Icons.timer_outlined,
                  label: 'Trip duration',
                  value: '${offer.estimatedTripDurationMin.round()} min',
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
