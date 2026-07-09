import 'package:flutter/material.dart';

import 'package:rider/features/booking/presentation/widgets/trip_point_cards.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

import '../../domain/models/driver_profile.dart';
import '../../domain/models/rider_trip_status.dart';
import '../widgets/contact_driver_button.dart';
import '../widgets/driver_info_card.dart';
import '../widgets/emergency_button.dart';
import '../widgets/trip_progress_indicator.dart';
import '../widgets/trip_status_banner.dart';

/// "Trip In Progress" — the trip has started. No Cancel Ride button here:
/// cancellation is only offered before the trip actually starts (see
/// `RiderTripStatus.isCancellable`).
class TripInProgressView extends StatelessWidget {
  const TripInProgressView({
    super.key,
    required this.tripSelection,
    required this.driver,
  });

  final TripSelection tripSelection;
  final DriverProfile driver;

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
        const TripStatusBanner(status: RiderTripStatus.inProgress),
        const SizedBox(height: 16),
        const TripProgressIndicator(status: RiderTripStatus.inProgress),
        const SizedBox(height: 16),
        DriverInfoCard(driver: driver),
        const SizedBox(height: 20),
        const Row(
          children: [
            Expanded(child: ContactDriverButton()),
            SizedBox(width: 12),
            Expanded(child: EmergencyButton()),
          ],
        ),
      ],
    );
  }
}
