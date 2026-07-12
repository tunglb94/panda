import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/booking/presentation/widgets/trip_point_cards.dart';
import 'package:rider/features/contact/domain/models/contact_info.dart';
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
    this.contact,
    required this.tripId,
    required this.apiClient,
    this.chatUnreadCount = 0,
  });

  final TripSelection tripSelection;
  final DriverProfile driver;
  final ContactInfo? contact;
  final String tripId;
  final ApiClient apiClient;
  final int chatUnreadCount;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        PickupCard(
          address: tripSelection.pickupAddress,
          coordinate: tripSelection.pickup,
        ),
        const RouteConnector(),
        DestinationCard(
          address: tripSelection.destinationAddress,
          coordinate: tripSelection.destination,
        ),
        const SizedBox(height: 20),
        const TripStatusBanner(status: RiderTripStatus.inProgress),
        const SizedBox(height: 16),
        const TripProgressIndicator(status: RiderTripStatus.inProgress),
        const SizedBox(height: 16),
        DriverInfoCard(driver: driver, contact: contact),
        const SizedBox(height: 20),
        Row(
          children: [
            Expanded(
              child: ContactDriverButton(tripId: tripId, apiClient: apiClient, unreadCount: chatUnreadCount),
            ),
            const SizedBox(width: 12),
            const Expanded(child: EmergencyButton()),
          ],
        ),
      ],
    );
  }
}
