import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';

/// Vehicle Statistics grid — "Chuyến đi" is the only real figure (reused
/// all-time trip count, same source as Profile's Statistics section — see
/// `EarningsRepository.fetchAllTimeTripCounts`). Km/Fuel/Service/Cost have
/// no backend source anywhere (no odometer tracking, no fuel log, no
/// service-cost ledger), so each shows an honest "—".
class VehicleStatsGrid extends StatelessWidget {
  const VehicleStatsGrid({super.key, required this.totalTrips});

  final int totalTrips;

  @override
  Widget build(BuildContext context) {
    final tiles = [
      _Tile(icon: Icons.route_outlined, label: 'Chuyến đi', value: '$totalTrips', isReal: true),
      const _Tile(icon: Icons.speed_outlined, label: 'Quãng đường', value: '—', isReal: false),
      const _Tile(icon: Icons.local_gas_station_outlined, label: 'Nhiên liệu', value: '—', isReal: false),
      const _Tile(icon: Icons.build_outlined, label: 'Lần bảo dưỡng', value: '—', isReal: false),
      const _Tile(icon: Icons.payments_outlined, label: 'Chi phí bảo trì', value: '—', isReal: false),
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
      itemBuilder: (context, i) => tiles[i],
    );
  }
}

class _Tile extends StatelessWidget {
  const _Tile({required this.icon, required this.label, required this.value, required this.isReal});

  final IconData icon;
  final String label;
  final String value;
  final bool isReal;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      animateIn: false,
      padding: const EdgeInsets.all(AppSpacing.md),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, color: isReal ? AppColors.primary : AppColors.textTertiary, size: AppIconSize.lg),
          const SizedBox(height: AppSpacing.xs),
          Text(
            value,
            style: Theme.of(context).textTheme.titleLarge?.copyWith(
                  color: isReal ? AppColors.textPrimary : AppColors.textTertiary,
                ),
          ),
          Text(label, style: Theme.of(context).textTheme.bodySmall),
        ],
      ),
    );
  }
}
