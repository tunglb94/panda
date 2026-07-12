import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_dialog.dart';
import '../../../../shared/widgets/app_settings_tile.dart';
import '../widgets/emergency_card.dart';
import '../widgets/incident_report_sheet.dart';
import '../widgets/safety_checklist.dart';
import 'trusted_contacts_page.dart';

/// Safety Center — a standalone module (per "Thiết kế module riêng"),
/// separate from Vehicle Center, bringing together everything a driver
/// needs in or before an unsafe situation: emergency numbers, trusted
/// contacts, live-location sharing, incident reporting, and a static
/// safety checklist. Every action here is either a real navigation (Trusted
/// Contacts, Incident Report, Support) or an honest, clearly-labeled
/// placeholder — nothing dials a real number or shares a real location.
class SafetyCenterPage extends StatelessWidget {
  const SafetyCenterPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Trung tâm an toàn')),
      body: ListView(
        padding: const EdgeInsets.all(AppSpacing.lg),
        children: [
          const _SafetyHeader(),
          const SizedBox(height: AppSpacing.lg),
          const EmergencyCard(),
          const SizedBox(height: AppSpacing.lg),
          AppCard(
            padding: EdgeInsets.zero,
            child: Column(
              children: [
                AppSettingsTile(
                  icon: Icons.share_location_outlined,
                  label: 'Chia sẻ vị trí',
                  subtitle: 'Chưa khả dụng — sẽ ra mắt trong giai đoạn tiếp theo',
                  onTap: () => AppDialog.info(
                    context,
                    title: 'Chia sẻ vị trí',
                    message: 'Tính năng chia sẻ vị trí thời gian thực với người thân '
                        'chưa khả dụng — sẽ ra mắt trong giai đoạn tiếp theo.',
                  ),
                ),
                const Divider(height: 1),
                AppSettingsTile(
                  icon: Icons.contacts_outlined,
                  label: 'Liên hệ tin cậy',
                  subtitle: 'Thêm người thân để liên hệ khi cần',
                  onTap: () => Navigator.of(context).push(
                    MaterialPageRoute(builder: (_) => const TrustedContactsPage()),
                  ),
                ),
                const Divider(height: 1),
                AppSettingsTile(
                  icon: Icons.report_gmailerrorred_outlined,
                  label: 'Báo cáo sự cố',
                  subtitle: 'Ghi lại sự việc xảy ra trong chuyến đi',
                  onTap: () => IncidentReportSheet.show(context),
                ),
              ],
            ),
          ),
          const SizedBox(height: AppSpacing.lg),
          const SafetyChecklist(),
        ],
      ),
    );
  }
}

class _SafetyHeader extends StatelessWidget {
  const _SafetyHeader();

  @override
  Widget build(BuildContext context) {
    return AppCard(
      color: AppColors.primaryLight,
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Row(
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: const BoxDecoration(color: AppColors.surface, shape: BoxShape.circle),
            child: const Icon(Icons.shield_outlined, color: AppColors.primary, size: AppIconSize.lg),
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'An toàn của bạn là ưu tiên hàng đầu',
                  style: Theme.of(context).textTheme.titleSmall?.copyWith(fontWeight: FontWeight.w700),
                ),
                const SizedBox(height: 2),
                Text(
                  'Truy cập nhanh số khẩn cấp, liên hệ tin cậy và mẹo an toàn khi lái xe.',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
