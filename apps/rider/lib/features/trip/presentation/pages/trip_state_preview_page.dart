import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/booking/domain/models/mock_booking_catalog.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/shared/utils/currency_format.dart';

import '../../domain/models/driver_profile.dart';
import '../../domain/models/rider_trip_status.dart';
import 'driver_arriving_view.dart';
import 'driver_assigned_view.dart';
import 'searching_driver_view.dart';
import 'trip_cancelled_view.dart';
import 'trip_completed_view.dart';
import 'trip_in_progress_view.dart';

/// Renders a single [RiderTripStatus] view in isolation, wrapped in its own
/// `Scaffold`, using sample data from the Booking module. Reachable from
/// `TripPreviewMenuPage` without needing to run the full lifecycle.
///
/// [apiClient] is only here so this preview's embedded `ContactDriverButton`
/// type-checks — [_previewTripId] is not a real trip, so tapping Call/Chat
/// in this dev-only screen will surface a real (harmless) API error rather
/// than doing anything destructive.
class TripStatePreviewPage extends StatelessWidget {
  const TripStatePreviewPage({super.key, required this.status, required this.apiClient});

  final RiderTripStatus status;
  final ApiClient apiClient;

  static const TripSelection _sampleTrip = MockBookingCatalog.sampleTripSelection;
  static const String _previewTripId = 'preview-trip';

  @override
  Widget build(BuildContext context) {
    // Dev-only preview screen (see class doc comment) — a fixed sample
    // amount is fine here; this never reaches a real rider and isn't the
    // real booking flow, which always quotes via PricingRepository.
    final fareText = formatMoney(45000, 'VND');
    const driver = DriverProfile(
      vehicleBrand: 'Toyota',
      vehicleModel: 'Vios',
      vehicleColor: 'Trắng',
      plateNumber: '51G-123.45',
    );

    return Scaffold(
      appBar: AppBar(title: Text('${status.label} (Xem trước)')),
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
                    tripId: _previewTripId,
                    apiClient: apiClient,
                    onCancel: () => Navigator.of(context).pop(),
                  ),
                RiderTripStatus.driverArriving => DriverArrivingView(
                    tripSelection: _sampleTrip,
                    driver: driver,
                    tripId: _previewTripId,
                    apiClient: apiClient,
                    onCancel: () => Navigator.of(context).pop(),
                  ),
                RiderTripStatus.inProgress => TripInProgressView(
                    tripSelection: _sampleTrip,
                    driver: driver,
                    tripId: _previewTripId,
                    apiClient: apiClient,
                  ),
                RiderTripStatus.completed ||
                RiderTripStatus.paymentPending ||
                RiderTripStatus.paymentSuccess ||
                RiderTripStatus.settled =>
                  TripCompletedView(
                    tripSelection: _sampleTrip,
                    driver: driver,
                    fareText: fareText,
                    onDone: () => Navigator.of(context).pop(),
                  ),
                RiderTripStatus.cancelled => TripCancelledView(
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
