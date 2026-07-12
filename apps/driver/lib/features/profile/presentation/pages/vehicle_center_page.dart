import 'package:flutter/material.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_bottom_sheet.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_empty_state.dart';
import '../../../../shared/widgets/app_skeleton.dart';
import '../../../../shared/widgets/app_status_chip.dart';
import '../../../earnings/data/earnings_repository.dart';
import '../../../vehicle/presentation/widgets/maintenance_section.dart';
import '../../../vehicle/presentation/widgets/vehicle_stats_grid.dart';
import '../../data/driver_profile_repository.dart';
import '../../domain/models/driver_own_profile.dart';
import '../widgets/info_row.dart';
import 'documents_page.dart';

class _VehicleCenterData {
  const _VehicleCenterData({required this.profile, required this.totalTrips});
  final DriverOwnProfile profile;
  final int totalTrips;
}

/// Vehicle Center — full redesign, replacing the old bare "Vehicle Summary"
/// page. Real fields (brand/model/plate/color/type/verification/total
/// trips) come from endpoints already used elsewhere in the app; the photo,
/// production year, per-document status, maintenance schedule, and
/// km/fuel/service stats have no backend source anywhere in this project
/// and are shown as honest, clearly-labeled placeholders rather than
/// invented values — see each section's doc comment for specifics.
class VehicleCenterPage extends StatefulWidget {
  const VehicleCenterPage({super.key, required this.apiClient, required this.driverId});

  final ApiClient apiClient;
  final String driverId;

  @override
  State<VehicleCenterPage> createState() => _VehicleCenterPageState();
}

class _VehicleCenterPageState extends State<VehicleCenterPage> {
  late Future<_VehicleCenterData> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<_VehicleCenterData> _load() async {
    final profileRepo = DriverProfileRepository(widget.apiClient);
    final earningsRepo = EarningsRepository(widget.apiClient);
    final results = await Future.wait([
      profileRepo.fetchOwnProfile(widget.driverId),
      earningsRepo.fetchAllTimeTripCounts(),
    ]);
    final counts = results[1] as (int, int);
    return _VehicleCenterData(
      profile: results[0] as DriverOwnProfile,
      totalTrips: counts.$1 + counts.$2,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Trung tâm xe')),
      body: FutureBuilder<_VehicleCenterData>(
        future: _future,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return ListView(
              padding: const EdgeInsets.all(AppSpacing.lg),
              children: const [
                AppSkeletonBox(height: 240),
                SizedBox(height: AppSpacing.lg),
                AppSkeletonBox(height: 200),
                SizedBox(height: AppSpacing.lg),
                AppSkeletonBox(height: 260),
              ],
            );
          }
          if (snap.hasError) {
            return AppEmptyState.error(
              subtitle: snap.error is ApiException && (snap.error as ApiException).statusCode == 0
                  ? (snap.error as ApiException).message
                  : 'Không thể tải thông tin xe.',
              onAction: () => setState(() => _future = _load()),
            );
          }

          final v = snap.data!.profile;
          final hasVehicle = v.vehicleBrand.isNotEmpty || v.vehicleModel.isNotEmpty;

          return RefreshIndicator(
            onRefresh: () async => setState(() => _future = _load()),
            child: ListView(
              padding: const EdgeInsets.all(AppSpacing.lg),
              children: [
                _VehiclePhotoCard(vehicle: v, hasVehicle: hasVehicle),
                const SizedBox(height: AppSpacing.lg),
                AppCard(
                  padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg),
                  child: Column(
                    children: [
                      InfoRow(label: 'Hãng xe', value: v.vehicleBrand.isEmpty ? 'Chưa cập nhật' : v.vehicleBrand),
                      const Divider(height: 1),
                      InfoRow(label: 'Model', value: v.vehicleModel.isEmpty ? 'Chưa cập nhật' : v.vehicleModel),
                      const Divider(height: 1),
                      InfoRow(label: 'Biển số', value: v.plateNumber.isEmpty ? 'Chưa cập nhật' : v.plateNumber),
                      const Divider(height: 1),
                      InfoRow(label: 'Loại xe', value: v.vehicleType.isEmpty ? 'Chưa cập nhật' : v.vehicleType),
                      const Divider(height: 1),
                      InfoRow(label: 'Màu xe', value: v.vehicleColor.isEmpty ? 'Chưa cập nhật' : v.vehicleColor),
                      const Divider(height: 1),
                      const InfoRow(label: 'Năm sản xuất', value: 'Chưa cập nhật'),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.xxl),
                Text('Giấy tờ xe', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: AppSpacing.md),
                _VehicleStatusCard(onViewAll: () => Navigator.of(context).push(
                      MaterialPageRoute(builder: (_) => const DocumentsPage()),
                    )),
                const SizedBox(height: AppSpacing.xxl),
                const MaintenanceSection(),
                const SizedBox(height: AppSpacing.xxl),
                Text('Thống kê xe', style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: AppSpacing.md),
                VehicleStatsGrid(totalTrips: snap.data!.totalTrips),
              ],
            ),
          );
        },
      ),
    );
  }
}

class _VehiclePhotoCard extends StatelessWidget {
  const _VehiclePhotoCard({required this.vehicle, required this.hasVehicle});

  final DriverOwnProfile vehicle;
  final bool hasVehicle;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      padding: EdgeInsets.zero,
      child: Column(
        children: [
          // No photo-upload endpoint exists — an honest empty photo slot
          // rather than a stock/placeholder car image pretending to be real.
          InkWell(
            borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
            onTap: () => _showPhotoPlaceholder(context),
            child: Container(
              height: 140,
              width: double.infinity,
              decoration: const BoxDecoration(
                color: AppColors.surfaceAlt,
                borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
              ),
              child: const Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.directions_car, size: AppIconSize.xxl, color: AppColors.textTertiary),
                  SizedBox(height: AppSpacing.xs),
                  Text('Chưa có ảnh xe', style: TextStyle(color: AppColors.textTertiary, fontSize: 12)),
                  SizedBox(height: 2),
                  Text('Nhấn để thêm ảnh', style: TextStyle(color: AppColors.primary, fontSize: 11, fontWeight: FontWeight.w600)),
                ],
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(AppSpacing.lg),
            child: Column(
              children: [
                Text(
                  hasVehicle ? vehicle.vehicleDisplay : 'Chưa cập nhật',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 4),
                Text(
                  vehicle.plateNumber.isEmpty ? 'Chưa cập nhật biển số' : vehicle.plateNumber,
                  style: Theme.of(context).textTheme.bodySmall,
                ),
                const SizedBox(height: AppSpacing.sm),
                AppStatusChip(
                  label: vehicle.isVerified ? 'Đã xác minh' : 'Chưa xác minh',
                  color: vehicle.isVerified ? AppColors.primary : AppColors.warning,
                  icon: vehicle.isVerified ? Icons.verified : Icons.hourglass_empty,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  void _showPhotoPlaceholder(BuildContext context) {
    AppBottomSheet.show<void>(
      context,
      title: 'Ảnh xe',
      builder: (sheetContext) => Text(
        'Tính năng tải ảnh xe chưa khả dụng — sẽ ra mắt trong giai đoạn tiếp theo.',
        style: Theme.of(sheetContext).textTheme.bodyMedium,
      ),
    );
  }
}

class _VehicleStatusCard extends StatelessWidget {
  const _VehicleStatusCard({required this.onViewAll});

  final VoidCallback onViewAll;

  static const _statuses = [
    'Đăng kiểm',
    'Bảo hiểm',
    'GPLX',
    'Đăng ký xe',
    'Giấy phép kinh doanh',
  ];

  @override
  Widget build(BuildContext context) {
    return AppCard(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Wrap(
            spacing: AppSpacing.sm,
            runSpacing: AppSpacing.sm,
            children: _statuses
                .map((s) => AppStatusChip(label: s, color: AppColors.textTertiary, icon: Icons.hourglass_empty))
                .toList(),
          ),
          const SizedBox(height: AppSpacing.md),
          Text(
            'Chưa có giấy tờ nào được xác minh.',
            style: Theme.of(context).textTheme.bodySmall,
          ),
          const SizedBox(height: AppSpacing.md),
          AppButton.outline(label: 'Xem chi tiết giấy tờ', onPressed: onViewAll, expand: true),
        ],
      ),
    );
  }
}
