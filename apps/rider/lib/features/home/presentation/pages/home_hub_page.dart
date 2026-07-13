import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/router/app_router.dart';
import 'package:rider/core/routing/route_provider.dart';
import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/booking/domain/models/vehicle_option.dart';
import 'package:rider/features/delivery/presentation/pages/delivery_form_page.dart';
import 'package:rider/features/history/presentation/pages/trip_history_page.dart';
import 'package:rider/features/map/presentation/pages/map_page.dart';
import 'package:rider/shared/widgets/app_card.dart';
import 'package:rider/shared/widgets/app_snackbar.dart';

/// The "Trang chủ" tab's content — a service hub (search bar + primary
/// service cards + a grid of secondary services), styled after Be/Xanh
/// SM/GSM's home screens. Booking itself (map, pickup/destination, vehicle
/// selection) still lives in [MapPage] and [DeliveryFormPage] exactly as
/// before — this page only decides which of those to open, matching the
/// reference apps' "hub picks the service, then a dedicated flow opens"
/// pattern instead of dropping straight into the map on every app launch.
///
/// Bike/Car are surfaced directly as quick picks (real vehicle artwork, see
/// `assets/vehicles/`) so a rider can jump straight into booking a specific
/// tier from Home instead of always landing on the generic map flow. "Đặt đồ
/// ăn" is a demo-only tile — Panda has no food-ordering feature yet; it just
/// shows what's coming next (per product ask), not a real entry point.
class HomeHubPage extends StatelessWidget {
  const HomeHubPage({super.key, required this.apiClient, required this.routeProvider});

  final ApiClient apiClient;
  final RouteProvider routeProvider;

  void _openMap(BuildContext context, {VehicleCategory? category}) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => MapPage(
          apiClient: apiClient,
          routeProvider: routeProvider,
          initialVehicleCategory: category,
        ),
      ),
    );
  }

  void _openDelivery(BuildContext context) {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => DeliveryFormPage(apiClient: apiClient)),
    );
  }

  void _openHistory(BuildContext context) {
    Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => TripHistoryPage(apiClient: apiClient)),
    );
  }

  void _showComingSoon(BuildContext context, String label) {
    AppSnackbar.show(context, '$label chưa khả dụng — sẽ ra mắt trong giai đoạn tiếp theo.');
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(AppSpacing.lg),
          children: [
            _SearchBar(onTap: () => _openMap(context)),
            const SizedBox(height: AppSpacing.lg),
            Row(
              children: [
                Expanded(
                  child: _PrimaryServiceCard(
                    imageAsset: 'assets/vehicles/bike.png',
                    label: 'Bike',
                    color: AppColors.primary,
                    onTap: () => _openMap(context, category: VehicleCategory.motorcycle),
                  ),
                ),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: _PrimaryServiceCard(
                    imageAsset: 'assets/vehicles/car.png',
                    label: 'Car',
                    color: AppColors.primary,
                    onTap: () => _openMap(context, category: VehicleCategory.car),
                  ),
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.md),
            Row(
              children: [
                Expanded(
                  child: _PrimaryServiceCard(
                    imageAsset: 'assets/services/send_package.png',
                    label: 'Gửi hàng',
                    color: AppColors.info,
                    onTap: () => _openDelivery(context),
                  ),
                ),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: _PrimaryServiceCard(
                    imageAsset: 'assets/services/order_food.png',
                    label: 'Đặt đồ ăn',
                    color: AppColors.info,
                    onTap: () => _showComingSoon(context, 'Đặt đồ ăn'),
                  ),
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.xl),
            Text('Dịch vụ khác', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: AppSpacing.md),
            Row(
              children: [
                Expanded(
                  child: _ServiceTile(
                    icon: Icons.account_balance_wallet_outlined,
                    label: 'Ví',
                    onTap: () => context.go(AppRoutes.wallet),
                  ),
                ),
                Expanded(
                  child: _ServiceTile(
                    icon: Icons.history,
                    label: 'Lịch sử',
                    onTap: () => _openHistory(context),
                  ),
                ),
                Expanded(
                  child: _ServiceTile(
                    icon: Icons.person_outline,
                    label: 'Hồ sơ',
                    onTap: () => context.go(AppRoutes.profile),
                  ),
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.xl),
            const _HomeBanner(),
          ],
        ),
      ),
    );
  }
}

class _SearchBar extends StatelessWidget {
  const _SearchBar({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: AppRadius.pillAll,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.md),
        decoration: BoxDecoration(
          color: AppColors.surfaceAlt,
          borderRadius: AppRadius.pillAll,
          border: Border.all(color: AppColors.border),
        ),
        child: Row(
          children: [
            Icon(Icons.search, size: AppIconSize.md, color: AppColors.textSecondary),
            const SizedBox(width: AppSpacing.sm),
            Text('Bạn muốn đi đâu?', style: Theme.of(context).textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary)),
          ],
        ),
      ),
    );
  }
}

class _PrimaryServiceCard extends StatelessWidget {
  const _PrimaryServiceCard({required this.imageAsset, required this.label, required this.color, required this.onTap});

  final String imageAsset;
  final String label;
  final Color color;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      onTap: onTap,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(color: color.withValues(alpha: 0.12), borderRadius: AppRadius.mdAll),
            child: Padding(
              padding: const EdgeInsets.all(4),
              child: Image.asset(imageAsset, fit: BoxFit.contain),
            ),
          ),
          const SizedBox(height: AppSpacing.sm),
          Text(label, style: Theme.of(context).textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w700)),
        ],
      ),
    );
  }
}

class _ServiceTile extends StatelessWidget {
  const _ServiceTile({required this.icon, required this.label, required this.onTap});

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: AppRadius.mdAll,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(color: AppColors.surfaceAlt, borderRadius: AppRadius.mdAll),
            child: Icon(icon, color: AppColors.textPrimary, size: AppIconSize.md),
          ),
          const SizedBox(height: AppSpacing.xs),
          Text(
            label,
            style: Theme.of(context).textTheme.bodySmall,
            textAlign: TextAlign.center,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ),
    );
  }
}

/// Fills the empty space at the bottom of Home with a promo banner. Purely
/// decorative today — no tap action, since nothing it could link to exists
/// yet (no promotions feature).
class _HomeBanner extends StatelessWidget {
  const _HomeBanner();

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: AppRadius.lgAll,
      child: Image.asset(
        'assets/banners/home_banner.jpg',
        width: double.infinity,
        fit: BoxFit.cover,
      ),
    );
  }
}
