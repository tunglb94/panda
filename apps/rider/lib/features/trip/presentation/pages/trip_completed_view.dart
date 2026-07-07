import 'package:flutter/material.dart';

import 'package:rider/features/booking/domain/models/mock_fare_calculator.dart';
import 'package:rider/features/booking/presentation/widgets/fare_summary_card.dart';
import 'package:rider/features/booking/presentation/widgets/trip_point_cards.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

import '../../domain/models/mock_driver.dart';
import '../../domain/models/rider_trip_status.dart';
import '../widgets/driver_info_card.dart';
import '../widgets/trip_progress_indicator.dart';
import '../widgets/trip_status_banner.dart';

/// "Trip Completed" — reuses the Booking UI's `FareSummaryCard` to show the
/// final fare. No Cancel/Contact/Emergency actions — the trip is over.
/// Rating the driver is out of scope here (Roadmap stage R7).
class TripCompletedView extends StatelessWidget {
  const TripCompletedView({
    super.key,
    required this.tripSelection,
    required this.driver,
    required this.fare,
    required this.onDone,
  });

  final TripSelection tripSelection;
  final MockDriver driver;
  final MockFareBreakdown fare;
  final VoidCallback onDone;

  @override
  Widget build(BuildContext context) {
    const status = RiderTripStatus.completed;
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
        const SizedBox(height: 16),
        FareSummaryCard(breakdown: fare),
        const SizedBox(height: 20),
        SizedBox(
          width: double.infinity,
          child: FilledButton(onPressed: onDone, child: const Text('Done')),
        ),
      ],
    );
  }
}
