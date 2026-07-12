import 'package:flutter/material.dart';

import 'package:rider/features/booking/presentation/widgets/trip_point_cards.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/shared/widgets/mascot_image.dart';

import '../../domain/models/rider_trip_status.dart';
import '../widgets/cancel_ride_button.dart';
import '../widgets/trip_progress_indicator.dart';
import '../widgets/trip_status_banner.dart';

/// "Searching Driver" — no driver has been matched yet.
///
/// Independently previewable: only needs a [TripSelection] and a cancel
/// callback, both supplied by the caller (either the live [MockTripRepository]
/// lifecycle or a dev preview screen).
class SearchingDriverView extends StatefulWidget {
  const SearchingDriverView({
    super.key,
    required this.tripSelection,
    required this.onCancel,
  });

  final TripSelection tripSelection;
  final VoidCallback onCancel;

  @override
  State<SearchingDriverView> createState() => _SearchingDriverViewState();
}

class _SearchingDriverViewState extends State<SearchingDriverView>
    with SingleTickerProviderStateMixin {
  late final AnimationController _pulseController;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 2),
    )..repeat(reverse: true);
  }

  @override
  void dispose() {
    _pulseController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    const status = RiderTripStatus.searchingDriver;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        PickupCard(
          address: widget.tripSelection.pickupAddress,
          coordinate: widget.tripSelection.pickup,
        ),
        const RouteConnector(),
        DestinationCard(
          address: widget.tripSelection.destinationAddress,
          coordinate: widget.tripSelection.destination,
        ),
        const SizedBox(height: 24),
        const TripStatusBanner(status: status),
        const SizedBox(height: 16),
        const TripProgressIndicator(status: status),
        const SizedBox(height: 32),
        Center(
          child: ScaleTransition(
            scale: Tween(begin: 0.94, end: 1.06).animate(
              CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
            ),
            // Reuses the existing pulse AnimationController rather than
            // adding a second animated element next to it — one cohesive
            // "waiting" motion instead of two competing ones.
            child: const MascotImage(
              asset: 'mascot_waiting.png',
              size: MascotSize.large,
              animation: MascotAnimation.none,
              semanticLabel: 'Đang tìm tài xế gần bạn',
            ),
          ),
        ),
        const SizedBox(height: 32),
        CancelRideButton(onCancel: widget.onCancel),
      ],
    );
  }
}
