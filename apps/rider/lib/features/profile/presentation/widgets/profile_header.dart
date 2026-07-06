import 'package:flutter/material.dart';

import '../../domain/models/rider_profile.dart';

/// Avatar (mock initials), full name, phone number, and member level badge.
class ProfileHeader extends StatelessWidget {
  const ProfileHeader({super.key, required this.profile});

  final RiderProfile profile;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    final level = profile.memberLevel;
    return Column(
      children: [
        CircleAvatar(
          radius: 44,
          backgroundColor: primary.withValues(alpha: 0.12),
          child: Text(
            profile.avatarInitial,
            style: TextStyle(fontSize: 32, fontWeight: FontWeight.bold, color: primary),
          ),
        ),
        const SizedBox(height: 12),
        Text(
          profile.fullName,
          style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
        ),
        const SizedBox(height: 4),
        Text(profile.phoneNumber, style: TextStyle(color: Colors.grey.shade500)),
        const SizedBox(height: 12),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          decoration: BoxDecoration(
            color: level.color.withValues(alpha: 0.12),
            borderRadius: BorderRadius.circular(20),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.workspace_premium, size: 16, color: level.color),
              const SizedBox(width: 4),
              Text(
                '${level.label} Member',
                style: TextStyle(color: level.color, fontWeight: FontWeight.w600),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
