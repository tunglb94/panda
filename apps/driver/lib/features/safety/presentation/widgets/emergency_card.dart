import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_dialog.dart';
import '../../../profile/presentation/pages/support_page.dart';

class _EmergencyEntry {
  const _EmergencyEntry({
    required this.icon,
    required this.label,
    this.number,
    this.color = AppColors.error,
  });

  final IconData icon;
  final String label;

  /// A real, publicly-known number (113/114/115) — `null` means PandaDriver
  /// has no number on file for this entry, shown as an honest placeholder
  /// rather than a made-up hotline.
  final String? number;
  final Color color;
}

const _entries = [
  _EmergencyEntry(icon: Icons.local_police_outlined, label: 'Công an', number: '113'),
  _EmergencyEntry(icon: Icons.local_fire_department_outlined, label: 'Cứu hỏa', number: '114'),
  _EmergencyEntry(icon: Icons.medical_services_outlined, label: 'Cấp cứu', number: '115'),
  _EmergencyEntry(
    icon: Icons.car_repair_outlined,
    label: 'Cứu hộ giao thông',
    color: AppColors.warning,
  ),
  _EmergencyEntry(
    icon: Icons.support_agent_outlined,
    label: 'Tổng đài Panda',
    color: AppColors.warning,
  ),
];

/// Emergency Card — real, publicly-known Vietnamese emergency numbers
/// (113/114/115) shown as accurate reference info, plus honest placeholders
/// for the numbers PandaDriver has no backend for (roadside assistance,
/// its own hotline). Per the explicit "Không gọi thật" requirement, no entry
/// here ever wires a real `tel:` call — tapping only informs the driver of
/// the number (if any) so they can dial it themselves from their phone app.
/// "Hỗ trợ" is the one genuine action: it opens the real, already-built
/// `SupportPage` rather than a dialog.
class EmergencyCard extends StatelessWidget {
  const EmergencyCard({super.key});

  @override
  Widget build(BuildContext context) {
    return AppCard(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.emergency_outlined, color: AppColors.error, size: AppIconSize.lg),
              const SizedBox(width: AppSpacing.sm),
              Expanded(
                child: Text('Số khẩn cấp', style: Theme.of(context).textTheme.titleMedium),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.xs),
          Text(
            'Ứng dụng chưa hỗ trợ gọi thật — chạm để xem số cần gọi.',
            style: Theme.of(context).textTheme.bodySmall,
          ),
          const SizedBox(height: AppSpacing.lg),
          GridView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _entries.length + 1,
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 3,
              mainAxisSpacing: AppSpacing.md,
              crossAxisSpacing: AppSpacing.sm,
              childAspectRatio: 0.85,
            ),
            itemBuilder: (context, i) {
              if (i == _entries.length) {
                return _EmergencyTile(
                  icon: Icons.headset_mic_outlined,
                  label: 'Hỗ trợ',
                  color: AppColors.primary,
                  onTap: () => Navigator.of(context).push(
                    MaterialPageRoute(builder: (_) => const SupportPage()),
                  ),
                );
              }
              final e = _entries[i];
              return _EmergencyTile(
                icon: e.icon,
                label: e.label,
                color: e.color,
                onTap: () => _showNumber(context, e),
              );
            },
          ),
        ],
      ),
    );
  }

  void _showNumber(BuildContext context, _EmergencyEntry e) {
    if (e.number == null) {
      AppDialog.info(
        context,
        title: e.label,
        message: 'PandaDriver chưa có số tổng đài cho mục này — sẽ ra mắt trong '
            'giai đoạn tiếp theo.',
      );
      return;
    }
    AppDialog.info(
      context,
      title: e.label,
      message: 'Số thật: ${e.number}. Ứng dụng chưa hỗ trợ gọi trực tiếp — nếu '
          'đây là tình huống khẩn cấp, vui lòng dùng ứng dụng điện thoại của '
          'bạn để gọi ${e.number}.',
      dismissLabel: 'Đã hiểu',
    );
  }
}

class _EmergencyTile extends StatelessWidget {
  const _EmergencyTile({
    required this.icon,
    required this.label,
    required this.color,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final Color color;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: AppColors.surfaceAlt,
      borderRadius: AppRadius.mdAll,
      child: InkWell(
        onTap: onTap,
        borderRadius: AppRadius.mdAll,
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(icon, color: color, size: AppIconSize.lg),
              const SizedBox(height: 6),
              Text(
                label,
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
