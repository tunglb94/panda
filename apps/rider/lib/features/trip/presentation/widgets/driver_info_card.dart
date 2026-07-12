import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/contact/domain/models/contact_info.dart';
import 'package:rider/shared/widgets/app_card.dart';

import '../../domain/models/driver_profile.dart';

class DriverInfoCard extends StatelessWidget {
  const DriverInfoCard({super.key, required this.driver, this.contact});

  final DriverProfile driver;

  /// Real name + rating (Part 4 — Contact Card), fetched separately via
  /// `ContactRepository.getContact` since the driver profile endpoint only
  /// ever returned vehicle data. Null while still loading — the card
  /// degrades to showing only the vehicle info it already had, never a
  /// fabricated name/rating.
  final ContactInfo? contact;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      child: Row(
        children: [
          Stack(
            clipBehavior: Clip.none,
            children: [
              Container(
                width: 52,
                height: 52,
                decoration: const BoxDecoration(
                  color: AppColors.primaryLight,
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.two_wheeler, color: AppColors.primary, size: AppIconSize.xl),
              ),
              if (driver.isVerified)
                Positioned(
                  right: -2,
                  bottom: -2,
                  child: Container(
                    padding: const EdgeInsets.all(2),
                    decoration: const BoxDecoration(
                      color: AppColors.surface,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.verified, size: AppIconSize.sm, color: AppColors.info),
                  ),
                ),
            ],
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (contact != null && contact!.name.isNotEmpty)
                  Text(
                    contact!.name,
                    style: Theme.of(context).textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w700),
                  ),
                Text(
                  driver.vehicleDisplay,
                  style: (contact != null && contact!.name.isNotEmpty)
                      ? Theme.of(context).textTheme.bodySmall
                      : Theme.of(context).textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w700),
                ),
                if (driver.vehicleColor.isNotEmpty) ...[
                  const SizedBox(height: 3),
                  Text(driver.vehicleColor, style: Theme.of(context).textTheme.bodySmall),
                ],
                if (contact != null && contact!.hasRating) ...[
                  const SizedBox(height: 3),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.star, size: 14, color: Colors.amber),
                      const SizedBox(width: 2),
                      Text(
                        contact!.rating.toStringAsFixed(1),
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(fontWeight: FontWeight.w600),
                      ),
                    ],
                  ),
                ],
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
            decoration: BoxDecoration(
              color: AppColors.surfaceAlt,
              borderRadius: AppRadius.mdAll,
              border: Border.all(color: AppColors.border),
            ),
            child: Text(
              driver.plateNumber,
              style: Theme.of(context).textTheme.labelLarge?.copyWith(letterSpacing: 0.3),
            ),
          ),
        ],
      ),
    );
  }
}
