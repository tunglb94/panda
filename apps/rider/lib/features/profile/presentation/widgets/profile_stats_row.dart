import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_card.dart';

import '../../domain/models/rider_profile.dart';

/// Rating and total completed trips, side by side.
class ProfileStatsRow extends StatelessWidget {
  const ProfileStatsRow({super.key, required this.profile});

  final RiderProfile profile;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      color: AppColors.surfaceAlt,
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg, vertical: AppSpacing.md),
      child: Row(
        children: [
          Expanded(
            child: _StatColumn(
              icon: Icons.star,
              iconColor: Colors.amber.shade700,
              label: 'Đánh giá',
              value: profile.rating.toStringAsFixed(1),
            ),
          ),
          Container(width: 1, height: 36, color: AppColors.border),
          Expanded(
            child: _StatColumn(
              icon: Icons.route,
              iconColor: AppColors.primary,
              label: 'Chuyến đã hoàn thành',
              value: '${profile.totalCompletedTrips}',
            ),
          ),
        ],
      ),
    );
  }
}

class _StatColumn extends StatelessWidget {
  const _StatColumn({
    required this.icon,
    required this.iconColor,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final Color iconColor;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(icon, size: AppIconSize.sm, color: iconColor),
            const SizedBox(width: AppSpacing.xs),
            Text(value, style: Theme.of(context).textTheme.titleMedium),
          ],
        ),
        const SizedBox(height: 2),
        Text(label, style: Theme.of(context).textTheme.labelSmall),
      ],
    );
  }
}
