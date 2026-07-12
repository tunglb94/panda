import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/booking/presentation/widgets/trip_point_cards.dart';
import 'package:rider/features/contact/domain/models/contact_info.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

import '../../domain/models/driver_profile.dart';
import '../../domain/models/rider_trip_status.dart';
import '../widgets/cancel_ride_button.dart';
import '../widgets/contact_driver_button.dart';
import '../widgets/driver_info_card.dart';
import '../widgets/emergency_button.dart';
import '../widgets/trip_progress_indicator.dart';
import '../widgets/trip_status_banner.dart';

/// "Driver Arriving" — the driver is close to the pickup point.
class DriverArrivingView extends StatelessWidget {
  const DriverArrivingView({
    super.key,
    required this.tripSelection,
    required this.driver,
    this.contact,
    required this.tripId,
    required this.apiClient,
    this.chatUnreadCount = 0,
    required this.onCancel,
  });

  final TripSelection tripSelection;
  final DriverProfile driver;
  final ContactInfo? contact;
  final String tripId;
  final ApiClient apiClient;
  final int chatUnreadCount;
  final VoidCallback onCancel;

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
        const TripStatusBanner(status: RiderTripStatus.driverArriving),
        const SizedBox(height: 16),
        const TripProgressIndicator(status: RiderTripStatus.driverArriving),
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
        const SizedBox(height: 12),
        CancelRideButton(onCancel: onCancel),
      ],
    );
  }
}
