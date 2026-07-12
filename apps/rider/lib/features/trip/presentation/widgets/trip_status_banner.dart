import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_shadows.dart';
import 'package:rider/core/theme/app_spacing.dart';

import '../../domain/models/rider_trip_status.dart';

/// Tinted hero card showing the current trip status: icon, headline, and a
/// short supporting message. Cross-fades its content whenever [status]
/// changes so the transition between lifecycle stages feels continuous
/// rather than an abrupt swap. A bespoke gradient hero (not a flat
/// `AppCard`) by design — its glow-badge icon treatment mirrors `apps/
/// driver`'s `AppShadows.glow()` pattern for emphasized elements.
class TripStatusBanner extends StatelessWidget {
  const TripStatusBanner({super.key, required this.status});

  final RiderTripStatus status;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.lg, AppSpacing.lg, AppSpacing.xl),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            AppColors.primary.withValues(alpha: 0.10),
            AppColors.primary.withValues(alpha: 0.03),
          ],
        ),
        borderRadius: AppRadius.xlAll,
        border: Border.all(color: AppColors.primary.withValues(alpha: 0.14)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(AppSpacing.md),
            decoration: BoxDecoration(
              color: AppColors.primary,
              shape: BoxShape.circle,
              boxShadow: AppShadows.glow(AppColors.primary),
            ),
            child: Icon(status.icon, color: AppColors.textOnPrimary, size: AppIconSize.lg),
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: AnimatedSwitcher(
              duration: const Duration(milliseconds: 300),
              child: Column(
                key: ValueKey(status),
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(status.label, style: Theme.of(context).textTheme.titleLarge),
                  const SizedBox(height: 3),
                  Text(status.statusMessage, style: Theme.of(context).textTheme.bodySmall),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
