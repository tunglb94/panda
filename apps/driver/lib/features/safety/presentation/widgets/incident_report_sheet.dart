import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_bottom_sheet.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/app_snackbar.dart';

class _IncidentCategory {
  const _IncidentCategory(this.icon, this.label);
  final IconData icon;
  final String label;
}

const _categories = [
  _IncidentCategory(Icons.car_crash_outlined, 'Tai nạn'),
  _IncidentCategory(Icons.report_problem_outlined, 'Khách gây rối'),
  _IncidentCategory(Icons.remove_shopping_cart_outlined, 'Mất tài sản'),
  _IncidentCategory(Icons.emergency_outlined, 'Khẩn cấp'),
  _IncidentCategory(Icons.more_horiz, 'Khác'),
];

/// Incident Report — a category picker + free-text description in an
/// `AppBottomSheet`. There is no incident-report backend anywhere in this
/// project, so "submitting" honestly tells the driver the report was not
/// actually sent rather than pretending to file a real report.
abstract final class IncidentReportSheet {
  static Future<void> show(BuildContext context) {
    return AppBottomSheet.show<void>(
      context,
      title: 'Báo cáo sự cố',
      isScrollControlled: true,
      builder: (sheetContext) => const _IncidentReportForm(),
    );
  }
}

class _IncidentReportForm extends StatefulWidget {
  const _IncidentReportForm();

  @override
  State<_IncidentReportForm> createState() => _IncidentReportFormState();
}

class _IncidentReportFormState extends State<_IncidentReportForm> {
  int? _selected;
  final _descController = TextEditingController();
  bool _submitted = false;

  @override
  void dispose() {
    _descController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_submitted) {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: AppSpacing.xl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.mark_email_read_outlined, size: 40, color: AppColors.primary),
            const SizedBox(height: AppSpacing.md),
            Text('Đã ghi nhận nội dung', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: AppSpacing.sm),
            Text(
              'Tính năng gửi báo cáo tới hệ thống chưa khả dụng — nội dung này '
              'chưa được gửi đi thật. Sẽ ra mắt trong giai đoạn tiếp theo.',
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodySmall,
            ),
            const SizedBox(height: AppSpacing.lg),
            AppButton.outline(label: 'Đóng', onPressed: () => Navigator.pop(context)),
          ],
        ),
      );
    }

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Loại sự cố', style: Theme.of(context).textTheme.labelLarge),
        const SizedBox(height: AppSpacing.sm),
        Wrap(
          spacing: AppSpacing.sm,
          runSpacing: AppSpacing.sm,
          children: [
            for (var i = 0; i < _categories.length; i++)
              _CategoryChoice(
                category: _categories[i],
                selected: _selected == i,
                onTap: () => setState(() => _selected = i),
              ),
          ],
        ),
        const SizedBox(height: AppSpacing.lg),
        Text('Mô tả (không bắt buộc)', style: Theme.of(context).textTheme.labelLarge),
        const SizedBox(height: AppSpacing.sm),
        TextField(
          controller: _descController,
          maxLines: 4,
          decoration: InputDecoration(
            hintText: 'Mô tả ngắn gọn sự việc…',
            filled: true,
            fillColor: AppColors.surfaceAlt,
            border: OutlineInputBorder(
              borderRadius: AppRadius.mdAll,
              borderSide: BorderSide.none,
            ),
          ),
        ),
        const SizedBox(height: AppSpacing.lg),
        AppButton.primary(
          label: 'Gửi báo cáo',
          onPressed: _selected == null
              ? null
              : () {
                  setState(() => _submitted = true);
                  AppSnackbar.warning(context, 'Báo cáo chưa được gửi tới hệ thống thật.');
                },
        ),
      ],
    );
  }
}

class _CategoryChoice extends StatelessWidget {
  const _CategoryChoice({required this.category, required this.selected, required this.onTap});

  final _IncidentCategory category;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: selected ? AppColors.primaryLight : AppColors.surfaceAlt,
      borderRadius: AppRadius.pillAll,
      child: InkWell(
        onTap: onTap,
        borderRadius: AppRadius.pillAll,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                category.icon,
                size: 16,
                color: selected ? AppColors.primary : AppColors.textSecondary,
              ),
              const SizedBox(width: 6),
              Text(
                category.label,
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: selected ? AppColors.primary : AppColors.textSecondary,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
