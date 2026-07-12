import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_bottom_sheet.dart';
import '../../../../shared/widgets/app_card.dart';

class _MaintenanceType {
  const _MaintenanceType(this.icon, this.label);
  final IconData icon;
  final String label;
}

const _types = [
  _MaintenanceType(Icons.build_circle_outlined, 'Bảo dưỡng định kỳ'),
  _MaintenanceType(Icons.oil_barrel_outlined, 'Thay dầu'),
  _MaintenanceType(Icons.tire_repair_outlined, 'Lốp xe'),
  _MaintenanceType(Icons.battery_charging_full_outlined, 'Ắc quy'),
  _MaintenanceType(Icons.disc_full_outlined, 'Phanh'),
  _MaintenanceType(Icons.shield_outlined, 'Bảo hiểm'),
];

/// Maintenance — UI only, exactly as scoped. There is no service-history or
/// reminder-scheduling backend, so every row honestly reads "Chưa thiết lập
/// nhắc nhở" rather than a fabricated last-service date or due date. The
/// vertical connector gives it a timeline feel without claiming any of the
/// entries represent a real logged event.
class MaintenanceSection extends StatelessWidget {
  const MaintenanceSection({super.key});

  @override
  Widget build(BuildContext context) {
    return AppCard(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text('Lịch bảo trì', style: Theme.of(context).textTheme.titleMedium),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: AppColors.surfaceAlt,
                  borderRadius: AppRadius.pillAll,
                ),
                child: const Text(
                  'Chưa có nhắc nhở',
                  style: TextStyle(fontSize: 10, color: AppColors.textSecondary, fontWeight: FontWeight.w600),
                ),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.md),
          for (var i = 0; i < _types.length; i++)
            _MaintenanceRow(
              type: _types[i],
              isLast: i == _types.length - 1,
            ),
        ],
      ),
    );
  }
}

class _MaintenanceRow extends StatelessWidget {
  const _MaintenanceRow({required this.type, required this.isLast});

  final _MaintenanceType type;
  final bool isLast;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: () => _showReminderPlaceholder(context, type.label),
      borderRadius: AppRadius.smAll,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 6),
        child: IntrinsicHeight(
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Column(
                children: [
                  Container(
                    width: 10,
                    height: 10,
                    decoration: const BoxDecoration(color: AppColors.border, shape: BoxShape.circle),
                  ),
                  if (!isLast)
                    Expanded(
                      child: Container(width: 2, color: AppColors.divider),
                    ),
                ],
              ),
              const SizedBox(width: AppSpacing.md),
              Icon(type.icon, size: AppIconSize.md, color: AppColors.textSecondary),
              const SizedBox(width: AppSpacing.sm),
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.only(bottom: AppSpacing.md),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(type.label, style: Theme.of(context).textTheme.bodyMedium),
                      Text(
                        'Chưa thiết lập nhắc nhở',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                  ),
                ),
              ),
              const Icon(Icons.add_circle_outline, size: 18, color: AppColors.primary),
            ],
          ),
        ),
      ),
    );
  }

  void _showReminderPlaceholder(BuildContext context, String label) {
    AppBottomSheet.show<void>(
      context,
      title: 'Nhắc nhở: $label',
      builder: (sheetContext) => Text(
        'Tính năng đặt lịch nhắc bảo trì chưa khả dụng — sẽ ra mắt trong '
        'giai đoạn tiếp theo.',
        style: Theme.of(sheetContext).textTheme.bodyMedium,
      ),
    );
  }
}
