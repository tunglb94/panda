import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_dialog.dart';

import '../../domain/models/surge_info.dart';

/// "⚡ Giá đang thay đổi" chip. Tapping it explains why in plain language —
/// deliberately reassuring, never alarming (per the sprint brief: "Không
/// làm khách hoảng. Giải thích ngắn gọn.").
///
/// Renders nothing when [surge] is null — see [SurgeInfo]'s doc comment for
/// why that is always the case in the app today.
class SurgeIndicator extends StatelessWidget {
  const SurgeIndicator({super.key, required this.surge});

  final SurgeInfo? surge;

  @override
  Widget build(BuildContext context) {
    final info = surge;
    if (info == null) return const SizedBox.shrink();

    return InkWell(
      borderRadius: AppRadius.pillAll,
      onTap: () => AppDialog.info(
        context,
        title: '⚡ ${info.label}',
        message: info.explanation,
        dismissLabel: 'Đã hiểu',
      ),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: 6),
        decoration: BoxDecoration(
          color: AppColors.warning.withValues(alpha: 0.12),
          borderRadius: AppRadius.pillAll,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('⚡', style: TextStyle(fontSize: 14)),
            const SizedBox(width: 6),
            Text(
              info.label,
              style: Theme.of(context).textTheme.labelMedium?.copyWith(color: AppColors.warning),
            ),
            const SizedBox(width: 4),
            const Icon(Icons.info_outline, size: 14, color: AppColors.warning),
          ],
        ),
      ),
    );
  }
}
