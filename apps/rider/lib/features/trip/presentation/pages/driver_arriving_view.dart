import 'package:flutter/material.dart';

import 'package:rider/features/booking/presentation/widgets/trip_point_cards.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

import '../../domain/models/mock_driver.dart';
import '../../domain/models/mock_trip_catalog.dart';
import '../../domain/models/rider_trip_status.dart';
import '../widgets/cancel_ride_button.dart';
import '../widgets/contact_driver_button.dart';
import '../widgets/driver_info_card.dart';
import '../widgets/emergency_button.dart';
import '../widgets/eta_arrival_card.dart';
import '../widgets/trip_progress_indicator.dart';
import '../widgets/trip_status_banner.dart';

/// "Driver Arriving" — the driver is close to the pickup point.
class DriverArrivingView extends StatelessWidget {
  const DriverArrivingView({
    super.key,
    required this.tripSelection,
    required this.driver,
    required this.onCancel,
  });

  final TripSelection tripSelection;
  final MockDriver driver;
  final VoidCallback onCancel;

  @override
  Widget build(BuildContext context) {
    const status = RiderTripStatus.driverArriving;
    final eta = MockTripCatalog.etaFor(status);
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
        const TripStatusBanner(status: status),
        const SizedBox(height: 16),
        const TripProgressIndicator(status: status),
        const SizedBox(height: 16),
        DriverInfoCard(driver: driver),
        const SizedBox(height: 12),
        EtaArrivalCard(
          eta: eta,
          arrivalLabel: MockTripCatalog.estimatedArrivalLabel(eta),
        ),
        const SizedBox(height: 20),
        const Row(
          children: [
            Expanded(child: ContactDriverButton()),
            SizedBox(width: 12),
            Expanded(child: EmergencyButton()),
          ],
        ),
        const SizedBox(height: 12),
        CancelRideButton(onCancel: onCancel),
      ],
    );
  }
}
