import 'package:flutter/material.dart';

import 'package:rider/features/booking/presentation/widgets/trip_point_cards.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

import '../../domain/models/driver_profile.dart';
import '../../domain/models/rider_trip_status.dart';
import '../widgets/driver_info_card.dart';
import '../widgets/trip_progress_indicator.dart';
import '../widgets/trip_status_banner.dart';

/// "Trip Completed" — shows the final fare returned from the backend.
class TripCompletedView extends StatelessWidget {
  const TripCompletedView({
    super.key,
    required this.tripSelection,
    required this.driver,
    required this.fareText,
    required this.onDone,
  });

  final TripSelection tripSelection;
  final DriverProfile driver;
  final String fareText;
  final VoidCallback onDone;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        PickupCard(
          address: tripSelection.pickupAddress,
          coordinate: tripSelection.pickup,
        ),
        const SizedBox(height: 8),
        DestinationCard(
          address: tripSelection.destinationAddress,
          coordinate: tripSelection.destination,
        ),
        const SizedBox(height: 20),
        const TripStatusBanner(status: RiderTripStatus.completed),
        const SizedBox(height: 16),
        const TripProgressIndicator(status: RiderTripStatus.completed),
        const SizedBox(height: 16),
        DriverInfoCard(driver: driver),
        const SizedBox(height: 16),
        _FinalFareCard(fareText: fareText),
        const SizedBox(height: 20),
        SizedBox(
          width: double.infinity,
          child: FilledButton(onPressed: onDone, child: const Text('Done')),
        ),
      ],
    );
  }
}

class _FinalFareCard extends StatelessWidget {
  const _FinalFareCard({required this.fareText});

  final String fareText;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          const Text('Final fare',
              style: TextStyle(fontWeight: FontWeight.bold)),
          Text(
            fareText,
            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
          ),
        ],
      ),
    );
  }
}
