import 'package:flutter/material.dart';

import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_skeleton.dart';

/// Loading placeholder shaped like the real page layout (dashboard card,
/// wallet card, quick actions row, a couple of transaction rows) — shown
/// while the first `GET /api/v1/driver/trips` call is in flight, instead of
/// a bare centered spinner.
class EarningsPageSkeleton extends StatelessWidget {
  const EarningsPageSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(AppSpacing.lg),
      children: [
        const AppSkeletonBox(height: 220, borderRadius: AppRadius.lgAll),
        const SizedBox(height: AppSpacing.lg),
        const AppSkeletonBox(height: 200, borderRadius: AppRadius.lgAll),
        const SizedBox(height: AppSpacing.lg),
        Row(
          children: List.generate(
            5,
            (i) => Expanded(
              child: Padding(
                padding: EdgeInsets.only(right: i == 4 ? 0 : AppSpacing.sm),
                child: const AppSkeletonBox(height: 64, borderRadius: AppRadius.mdAll),
              ),
            ),
          ),
        ),
        const SizedBox(height: AppSpacing.lg),
        const AppSkeletonListTile(),
        const SizedBox(height: AppSpacing.sm),
        const AppSkeletonListTile(),
        const SizedBox(height: AppSpacing.sm),
        const AppSkeletonListTile(),
      ],
    );
  }
}
