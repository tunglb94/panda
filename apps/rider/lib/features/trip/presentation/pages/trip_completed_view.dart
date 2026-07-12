import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/booking/presentation/widgets/trip_point_cards.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/shared/widgets/mascot_image.dart';

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
        const Center(
          child: MascotImage(
            asset: 'mascot_celebration.png',
            size: MascotSize.medium,
            animation: MascotAnimation.bounce,
            semanticLabel: 'Chuyến đi hoàn tất',
          ),
        ),
        const SizedBox(height: AppSpacing.md),
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
          child: FilledButton(onPressed: onDone, child: const Text('Xong')),
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
    final theme = Theme.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg, vertical: AppSpacing.md),
      decoration: BoxDecoration(
        color: AppColors.surfaceAlt,
        borderRadius: AppRadius.mdAll,
        border: Border.all(color: AppColors.border),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text('Cước phí cuối cùng', style: theme.textTheme.titleSmall),
          Flexible(
            child: Text(
              fareText,
              textAlign: TextAlign.right,
              style: theme.textTheme.titleMedium?.copyWith(color: AppColors.primary),
            ),
          ),
        ],
      ),
    );
  }
}
