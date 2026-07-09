import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/booking/domain/models/mock_booking_catalog.dart';
import 'package:rider/features/booking/presentation/widgets/booking_form_body.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

/// Booking tab entry point.
///
/// If reached from the Map page's confirmed pickup/destination selection,
/// [tripSelection] carries the real selection over. Otherwise a sample
/// selection is used so this screen is always demoable on its own.
class BookingPage extends StatelessWidget {
  const BookingPage({
    super.key,
    this.tripSelection,
    required this.apiClient,
  });

  final TripSelection? tripSelection;
  final ApiClient apiClient;

  @override
  Widget build(BuildContext context) {
    final trip = tripSelection ?? MockBookingCatalog.sampleTripSelection;
    return Scaffold(
      appBar: AppBar(title: const Text('Book a Ride')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: BookingFormBody(
                tripSelection: trip,
                apiClient: apiClient,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
