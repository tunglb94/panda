import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';

import '../../domain/models/quick_reply.dart';

/// Horizontal row of canned-reply chips (Part 2). Wraps rather than scrolls
/// when there isn't enough width, so small screens never clip a chip.
class QuickReplyRow extends StatelessWidget {
  const QuickReplyRow({super.key, required this.tripType, required this.onSelect});

  final String tripType;
  final ValueChanged<String> onSelect;

  @override
  Widget build(BuildContext context) {
    final keys = quickRepliesFor(tripType);
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
      child: Wrap(
        spacing: AppSpacing.sm,
        runSpacing: AppSpacing.xs,
        children: [
          for (final key in keys)
            Semantics(
              button: true,
              label: 'Trả lời nhanh: ${quickReplyLabels[key]}',
              child: GestureDetector(
                onTap: () => onSelect(key),
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  decoration: BoxDecoration(
                    color: AppColors.surface,
                    borderRadius: AppRadius.pillAll,
                    border: Border.all(color: AppColors.border),
                  ),
                  child: Text(
                    quickReplyLabels[key] ?? key,
                    style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppColors.primary),
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
