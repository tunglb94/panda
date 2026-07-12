import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_radius.dart';
import '../../core/theme/app_spacing.dart';

/// Standard modal bottom sheet chrome: rounded top corners, drag handle,
/// optional title, safe-area padding. Mirrors `apps/driver`'s
/// `AppBottomSheet` exactly — wraps [showModalBottomSheet] so every sheet
/// in the app (booking form, place picker, safety actions) looks like the
/// same component instead of each screen hand-rolling its own chrome.
abstract final class AppBottomSheet {
  static Future<T?> show<T>(
    BuildContext context, {
    required WidgetBuilder builder,
    String? title,
    bool isScrollControlled = false,
  }) {
    return showModalBottomSheet<T>(
      context: context,
      isScrollControlled: isScrollControlled,
      backgroundColor: AppColors.surface,
      shape: const RoundedRectangleBorder(borderRadius: AppRadius.topXl),
      builder: (sheetContext) => SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(
            AppSpacing.lg,
            AppSpacing.md,
            AppSpacing.lg,
            AppSpacing.lg,
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  margin: const EdgeInsets.only(bottom: AppSpacing.lg),
                  decoration: BoxDecoration(
                    color: AppColors.border,
                    borderRadius: AppRadius.smAll,
                  ),
                ),
              ),
              if (title != null) ...[
                Text(title, style: Theme.of(sheetContext).textTheme.titleLarge),
                const SizedBox(height: AppSpacing.md),
              ],
              builder(sheetContext),
            ],
          ),
        ),
      ),
    );
  }
}
