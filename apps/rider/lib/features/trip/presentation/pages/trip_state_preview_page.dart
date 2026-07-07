import 'package:flutter/material.dart';

import 'package:rider/features/booking/domain/models/mock_booking_catalog.dart';
import 'package:rider/features/booking/domain/models/mock_fare_calculator.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

import '../../domain/models/mock_trip_catalog.dart';
import '../../domain/models/rider_trip_status.dart';
import 'driver_arriving_view.dart';
import 'driver_assigned_view.dart';
import 'searching_driver_view.dart';
import 'trip_completed_view.dart';
import 'trip_in_progress_view.dart';

/// Renders a single [RiderTripStatus] view in isolation, wrapped in its own
/// `Scaffold`, using sample data from the Booking module. This is the
/// "independently previewable for development" entry point requested for
/// Phase R-02 — reachable from `TripPreviewMenuPage` without needing to run
/// the full mock lifecycle.
class TripStatePreviewPage extends StatelessWidget {
  const TripStatePreviewPage({super.key, required this.status});

  final RiderTripStatus status;

  static const TripSelection _sampleTrip = MockBookingCatalog.sampleTripSelection;

  @override
  Widget build(BuildContext context) {
    final vehicle = MockBookingCatalog.vehicles.first;
    final fare = MockFareBreakdown.calculate(
      vehicle: vehicle,
      distanceKm: 6.4,
      durationMin: 14,
    );
    final driver = MockTripCatalog.sampleDriver;

    return Scaffold(
      appBar: AppBar(title: Text('${status.label} (Preview)')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: switch (status) {
                RiderTripStatus.searchingDriver => SearchingDriverView(
                    tripSelection: _sampleTrip,
                    onCancel: () => Navigator.of(context).pop(),
                  ),
                RiderTripStatus.driverAssigned => DriverAssignedView(
                    tripSelection: _sampleTrip,
                    driver: driver,
                    onCancel: () => Navigator.of(context).pop(),
                  ),
                RiderTripStatus.driverArriving => DriverArrivingView(
                    tripSelection: _sampleTrip,
                    driver: driver,
                    onCancel: () => Navigator.of(context).pop(),
                  ),
                RiderTripStatus.inProgress => TripInProgressView(
                    tripSelection: _sampleTrip,
                    driver: driver,
                  ),
                RiderTripStatus.completed => TripCompletedView(
                    tripSelection: _sampleTrip,
                    driver: driver,
                    fare: fare,
                    onDone: () => Navigator.of(context).pop(),
                  ),
              },
            ),
          ),
        ),
      ),
    );
  }
}
