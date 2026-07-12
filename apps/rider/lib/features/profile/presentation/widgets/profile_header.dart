import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_status_chip.dart';

import '../../domain/models/rider_profile.dart';

/// Avatar (mock initials), full name, phone number, and member level badge.
class ProfileHeader extends StatelessWidget {
  const ProfileHeader({super.key, required this.profile});

  final RiderProfile profile;

  @override
  Widget build(BuildContext context) {
    final level = profile.memberLevel;
    return Column(
      children: [
        CircleAvatar(
          radius: 44,
          backgroundColor: AppColors.primaryLight,
          child: Text(
            profile.avatarInitial,
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(color: AppColors.primary),
          ),
        ),
        const SizedBox(height: AppSpacing.md),
        Text(profile.fullName, style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: 4),
        Text(profile.phoneNumber, style: Theme.of(context).textTheme.bodySmall),
        const SizedBox(height: AppSpacing.md),
        AppStatusChip(
          label: '${level.label} Member',
          color: level.color,
          icon: Icons.workspace_premium,
        ),
      ],
    );
  }
}
