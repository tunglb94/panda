import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';

/// Statistics grid. Only "Chuyến đi" (trip counts) is computed from real
/// data (`GET /api/v1/driver/trips`, all-time). Acceptance/completion/
/// cancellation rate, rating, online hours, and distance have no backend
/// source anywhere — the trips list carries no accept/reject history, no
/// online-session log, no distance, no rating. Each shows an honest "—"
/// rather than a guess or a fabricated percentage.
class StatisticsGrid extends StatelessWidget {
  const StatisticsGrid({super.key, required this.completedTrips, required this.cancelledTrips});

  final int completedTrips;
  final int cancelledTrips;

  @override
  Widget build(BuildContext context) {
    final tiles = [
      _StatTileData(
        icon: Icons.route_outlined,
        label: 'Chuyến đi',
        value: '$completedTrips',
        color: AppColors.primary,
        isReal: true,
      ),
      const _StatTileData(
        icon: Icons.thumb_up_outlined,
        label: 'Tỉ lệ nhận chuyến',
        value: '—',
        color: AppColors.info,
        isReal: false,
      ),
      const _StatTileData(
        icon: Icons.task_alt,
        label: 'Tỉ lệ hoàn thành',
        value: '—',
        color: AppColors.primary,
        isReal: false,
      ),
      _StatTileData(
        icon: Icons.cancel_outlined,
        label: 'Đã hủy',
        value: '$cancelledTrips',
        color: AppColors.textTertiary,
        isReal: true,
      ),
      const _StatTileData(
        icon: Icons.star_outline,
        label: 'Đánh giá',
        value: '—',
        color: AppColors.warning,
        isReal: false,
      ),
      const _StatTileData(
        icon: Icons.schedule_outlined,
        label: 'Giờ hoạt động',
        value: '—',
        color: AppColors.info,
        isReal: false,
      ),
      const _StatTileData(
        icon: Icons.straighten_outlined,
        label: 'Quãng đường',
        value: '—',
        color: AppColors.info,
        isReal: false,
      ),
    ];

    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: tiles.length,
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        mainAxisSpacing: AppSpacing.md,
        crossAxisSpacing: AppSpacing.md,
        childAspectRatio: 1.6,
      ),
      itemBuilder: (context, i) => _StatTile(data: tiles[i]),
    );
  }
}

class _StatTileData {
  const _StatTileData({
    required this.icon,
    required this.label,
    required this.value,
    required this.color,
    required this.isReal,
  });

  final IconData icon;
  final String label;
  final String value;
  final Color color;
  final bool isReal;
}

class _StatTile extends StatelessWidget {
  const _StatTile({required this.data});

  final _StatTileData data;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      animateIn: false,
      padding: const EdgeInsets.all(AppSpacing.md),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(data.icon, color: data.color, size: AppIconSize.lg),
          const SizedBox(height: AppSpacing.xs),
          Text(
            data.value,
            style: Theme.of(context).textTheme.titleLarge?.copyWith(
                  color: data.isReal ? AppColors.textPrimary : AppColors.textTertiary,
                ),
          ),
          Text(data.label, style: Theme.of(context).textTheme.bodySmall),
        ],
      ),
    );
  }
}
