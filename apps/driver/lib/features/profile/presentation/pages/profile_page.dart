import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/auth/auth_state.dart';
import '../../../../core/network/api_client.dart';
import '../../../../core/router/app_router.dart';
import '../../../../core/storage/token_storage.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_empty_state.dart';
import '../../../../shared/widgets/app_skeleton.dart';
import '../../../../shared/widgets/app_snackbar.dart';
import '../../../../shared/widgets/app_status_chip.dart';
import '../../../../shared/widgets/pressable_scale.dart';
import '../../../earnings/data/earnings_repository.dart';
import '../../../earnings/presentation/widgets/driver_level_card.dart';
import '../../../earnings/presentation/widgets/statistics_grid.dart';
import '../../../home/data/availability_repository.dart';
import '../../../safety/presentation/pages/safety_center_page.dart';
import '../../../kyc/data/kyc_repository.dart';
import '../../../kyc/domain/models/kyc_status.dart';
import '../../../kyc/presentation/pages/kyc_status_page.dart';
import '../../../wallet/presentation/pages/wallet_page.dart';
import '../../data/driver_profile_repository.dart';
import '../../domain/models/driver_own_profile.dart';
import 'driver_trip_history_page.dart';
import 'settings_page.dart';
import 'support_page.dart';
import 'vehicle_center_page.dart';

class _ProfileData {
  const _ProfileData({
    required this.profile,
    required this.isOnline,
    required this.tripCounts,
    required this.kycApproved,
  });
  final DriverOwnProfile profile;
  final bool isOnline;
  final (int, int) tripCounts;

  /// Real Driver KYC + Vehicle Verification status (both Approved) — the
  /// richer, document-backed status this phase adds, distinct from
  /// [DriverOwnProfile.isVerified] (the older, coarse legacy field).
  final bool kycApproved;
}

/// Profile — header (verification/online status/trip count/join date are
/// all real; rating and display name are not available anywhere in the
/// backend and are shown/omitted honestly), reused `StatisticsGrid` and
/// `DriverLevelCard` from the Earnings sprint (per the explicit instruction
/// not to rebuild them), and Quick Shortcuts into every other profile-
/// adjacent screen.
class ProfilePage extends StatefulWidget {
  const ProfilePage({
    super.key,
    required this.apiClient,
    required this.authState,
    required this.tokenStorage,
  });

  final ApiClient apiClient;
  final AuthState authState;
  final TokenStorage tokenStorage;

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  late Future<_ProfileData> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<_ProfileData> _load() async {
    final driverId = widget.authState.driverId ?? '';
    final profileRepo = DriverProfileRepository(widget.apiClient);
    final availRepo = AvailabilityRepository(widget.apiClient);
    final earningsRepo = EarningsRepository(widget.apiClient);
    final kycRepo = KYCRepository(widget.apiClient);

    final results = await Future.wait([
      profileRepo.fetchOwnProfile(driverId),
      availRepo.getAvailability(),
      earningsRepo.fetchAllTimeTripCounts(),
    ]);

    // Best-effort, independent of the 3 calls above — a KYC fetch failure
    // (e.g. nothing submitted yet, 404) must never break the rest of Profile.
    var kycApproved = false;
    try {
      final driverV = await kycRepo.getDriverVerification();
      final vehicleV = await kycRepo.getVehicleVerification();
      kycApproved = (driverV?.status.isApproved ?? false) && (vehicleV?.status.isApproved ?? false);
    } catch (_) {
      // Ignore — badge just shows "Chưa xác minh".
    }

    return _ProfileData(
      profile: results[0] as DriverOwnProfile,
      isOnline: (results[1] as AvailabilityResult).isOnline,
      tripCounts: results[2] as (int, int),
      kycApproved: kycApproved,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Hồ sơ'),
        actions: [
          IconButton(
            icon: const Icon(Icons.settings_outlined),
            tooltip: 'Cài đặt',
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => SettingsPage(
                  authState: widget.authState,
                  tokenStorage: widget.tokenStorage,
                  driverId: widget.authState.driverId ?? '',
                  buildVehiclePage: () => VehicleCenterPage(
                    apiClient: widget.apiClient,
                    driverId: widget.authState.driverId ?? '',
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
      body: FutureBuilder<_ProfileData>(
        future: _future,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return const _ProfileSkeleton();
          }
          if (snap.hasError) {
            return AppEmptyState.error(
              subtitle: snap.error is ApiException && (snap.error as ApiException).statusCode == 0
                  ? (snap.error as ApiException).message
                  : 'Không thể tải hồ sơ.',
              onAction: () => setState(() => _future = _load()),
            );
          }

          final data = snap.data!;
          return RefreshIndicator(
            onRefresh: () async => setState(() => _future = _load()),
            child: ListView(
              padding: const EdgeInsets.all(AppSpacing.lg),
              children: [
                _ProfileHeader(data: data, apiClient: widget.apiClient),
                const SizedBox(height: AppSpacing.xl),
                Text('Thống kê', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: AppSpacing.md),
                StatisticsGrid(
                  completedTrips: data.tripCounts.$1,
                  cancelledTrips: data.tripCounts.$2,
                ),
                const SizedBox(height: AppSpacing.xxl),
                const DriverLevelCard(),
                const SizedBox(height: AppSpacing.xxl),
                Text('Lối tắt', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: AppSpacing.md),
                _QuickShortcutsGrid(apiClient: widget.apiClient, authState: widget.authState),
              ],
            ),
          );
        },
      ),
    );
  }
}

class _ProfileHeader extends StatelessWidget {
  const _ProfileHeader({required this.data, required this.apiClient});

  final _ProfileData data;
  final ApiClient apiClient;

  @override
  Widget build(BuildContext context) {
    final profile = data.profile;
    return AppCard(
      padding: const EdgeInsets.all(AppSpacing.xl),
      child: Column(
        children: [
          Stack(
            clipBehavior: Clip.none,
            children: [
              Container(
                width: 84,
                height: 84,
                decoration: const BoxDecoration(color: AppColors.primaryLight, shape: BoxShape.circle),
                child: const Icon(Icons.person, size: 44, color: AppColors.primary),
              ),
              Positioned(
                right: 0,
                bottom: 2,
                child: Container(
                  width: 20,
                  height: 20,
                  decoration: BoxDecoration(
                    color: data.isOnline ? AppColors.primary : AppColors.textTertiary,
                    shape: BoxShape.circle,
                    border: Border.all(color: AppColors.surface, width: 3),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.md),
          Text('Tài xế PandaDriver', style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 4),
          Text(
            data.isOnline ? 'Đang online' : 'Đang offline',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: data.isOnline ? AppColors.primary : AppColors.textTertiary,
                  fontWeight: FontWeight.w600,
                ),
          ),
          const SizedBox(height: AppSpacing.md),
          Wrap(
            spacing: AppSpacing.sm,
            runSpacing: AppSpacing.sm,
            alignment: WrapAlignment.center,
            children: [
              InkWell(
                borderRadius: BorderRadius.circular(999),
                onTap: () => Navigator.of(context).push(
                  MaterialPageRoute(builder: (_) => KYCStatusPage(apiClient: apiClient)),
                ),
                child: AppStatusChip(
                  label: data.kycApproved ? 'Đã xác minh' : 'Chưa xác minh — Xem chi tiết',
                  color: data.kycApproved ? AppColors.primary : AppColors.warning,
                  icon: data.kycApproved ? Icons.verified : Icons.hourglass_empty,
                ),
              ),
              const AppStatusChip(label: 'Đánh giá —', color: AppColors.textTertiary, icon: Icons.star_outline),
            ],
          ),
          const SizedBox(height: AppSpacing.lg),
          const Divider(height: 1),
          const SizedBox(height: AppSpacing.md),
          Row(
            children: [
              Expanded(
                child: _HeaderStat(
                  label: 'Tổng chuyến',
                  value: '${data.tripCounts.$1}',
                ),
              ),
              Container(width: 1, height: 32, color: AppColors.border),
              Expanded(
                child: _HeaderStat(
                  label: 'Gia nhập từ',
                  value: profile.createdAt != null
                      ? '${profile.createdAt!.month}/${profile.createdAt!.year}'
                      : '—',
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _HeaderStat extends StatelessWidget {
  const _HeaderStat({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(value, style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 2),
        Text(label, style: Theme.of(context).textTheme.bodySmall),
      ],
    );
  }
}

class _QuickShortcutsGrid extends StatelessWidget {
  const _QuickShortcutsGrid({required this.apiClient, required this.authState});

  final ApiClient apiClient;
  final AuthState authState;

  @override
  Widget build(BuildContext context) {
    final driverId = authState.driverId ?? '';
    final shortcuts = <(IconData, String, VoidCallback)>[
      (
        Icons.account_balance_wallet_outlined,
        'Ví',
        () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => WalletPage(apiClient: apiClient))),
      ),
      (Icons.bar_chart_outlined, 'Thu nhập', () => context.go(AppRoutes.earnings)),
      (
        Icons.receipt_long_outlined,
        'Lịch sử chuyến',
        () => Navigator.of(context).push(
              MaterialPageRoute(builder: (_) => DriverTripHistoryPage(apiClient: apiClient)),
            ),
      ),
      (
        Icons.directions_car_outlined,
        'Xe',
        () => Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => VehicleCenterPage(apiClient: apiClient, driverId: driverId),
              ),
            ),
      ),
      (
        Icons.account_balance_outlined,
        'Ngân hàng',
        () => AppSnackbar.show(context, 'Liên kết ngân hàng chưa khả dụng — sẽ ra mắt trong giai đoạn tiếp theo.'),
      ),
      (
        Icons.badge_outlined,
        'Xác minh',
        () => Navigator.of(context).push(
              MaterialPageRoute(builder: (_) => KYCStatusPage(apiClient: apiClient)),
            ),
      ),
      (Icons.notifications_outlined, 'Thông báo', () => context.go(AppRoutes.notifications)),
      (
        Icons.support_agent_outlined,
        'Hỗ trợ',
        () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const SupportPage())),
      ),
      (
        Icons.shield_outlined,
        'An toàn',
        () => Navigator.of(context).push(
              MaterialPageRoute(builder: (_) => const SafetyCenterPage()),
            ),
      ),
    ];

    return GridView.builder(
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: shortcuts.length,
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 4,
        mainAxisSpacing: AppSpacing.md,
        crossAxisSpacing: AppSpacing.sm,
        childAspectRatio: 0.8,
      ),
      itemBuilder: (context, i) {
        final (icon, label, onTap) = shortcuts[i];
        return _ShortcutButton(icon: icon, label: label, onTap: onTap);
      },
    );
  }
}

class _ShortcutButton extends StatefulWidget {
  const _ShortcutButton({required this.icon, required this.label, required this.onTap});

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  State<_ShortcutButton> createState() => _ShortcutButtonState();
}

class _ShortcutButtonState extends State<_ShortcutButton> {
  bool _pressed = false;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: widget.onTap,
        onHighlightChanged: (v) => setState(() => _pressed = v),
        borderRadius: BorderRadius.circular(14),
        child: PressableScale(
          pressed: _pressed,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 48,
                height: 48,
                decoration: const BoxDecoration(color: AppColors.primaryLight, shape: BoxShape.circle),
                child: Icon(widget.icon, color: AppColors.primary, size: AppIconSize.md),
              ),
              const SizedBox(height: 6),
              Text(
                widget.label,
                textAlign: TextAlign.center,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w600),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ProfileSkeleton extends StatelessWidget {
  const _ProfileSkeleton();

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(AppSpacing.lg),
      children: const [
        AppSkeletonBox(height: 220),
        SizedBox(height: AppSpacing.xl),
        AppSkeletonBox(height: 160),
        SizedBox(height: AppSpacing.xl),
        AppSkeletonBox(height: 260),
      ],
    );
  }
}
