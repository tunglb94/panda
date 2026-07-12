import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_bottom_sheet.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_status_chip.dart';

class _DocumentType {
  const _DocumentType(this.icon, this.title);
  final IconData icon;
  final String title;
}

const _documents = [
  _DocumentType(Icons.badge_outlined, 'Giấy phép lái xe (GPLX)'),
  _DocumentType(Icons.credit_card_outlined, 'Căn cước công dân (CCCD)'),
  _DocumentType(Icons.description_outlined, 'Đăng ký xe'),
  _DocumentType(Icons.shield_outlined, 'Bảo hiểm'),
  _DocumentType(Icons.fact_check_outlined, 'Kiểm định (Đăng kiểm)'),
  _DocumentType(Icons.storefront_outlined, 'Giấy phép kinh doanh'),
];

/// Documents — status-only, per spec. The backend has no per-document
/// verification tracking (only one overall `verification_status` on the
/// driver profile), so every document here honestly reads "Chưa cập nhật"
/// rather than a guessed per-document status. The upload action is a
/// placeholder — there is no document-upload endpoint.
class DocumentsPage extends StatelessWidget {
  const DocumentsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Giấy tờ')),
      body: ListView.separated(
        padding: const EdgeInsets.all(AppSpacing.lg),
        itemCount: _documents.length,
        separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
        itemBuilder: (context, i) {
          final doc = _documents[i];
          return AppCard(
            padding: const EdgeInsets.all(AppSpacing.md),
            onTap: () => _showUploadPlaceholder(context, doc.title),
            child: Row(
              children: [
                Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(color: AppColors.surfaceAlt, shape: BoxShape.circle),
                  child: Icon(doc.icon, color: AppColors.textSecondary, size: AppIconSize.md),
                ),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: Text(doc.title, style: Theme.of(context).textTheme.bodyMedium),
                ),
                const AppStatusChip(label: 'Chưa cập nhật', color: AppColors.textTertiary, icon: Icons.upload_outlined),
              ],
            ),
          );
        },
      ),
    );
  }

  void _showUploadPlaceholder(BuildContext context, String docTitle) {
    AppBottomSheet.show<void>(
      context,
      title: docTitle,
      builder: (sheetContext) => Text(
        'Tính năng tải lên giấy tờ chưa khả dụng — sẽ ra mắt trong '
        'giai đoạn tiếp theo.',
        style: Theme.of(sheetContext).textTheme.bodyMedium,
      ),
    );
  }
}
