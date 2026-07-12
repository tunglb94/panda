import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_spacing.dart';
import 'app_button.dart';

/// Standard confirmation dialog. Replaces one-off `showDialog` +
/// `AlertDialog` blocks (e.g. Cancel Ride, and the logout/end-trip
/// confirmations flagged in `docs/driver/DRIVER_APP_SPEC.md` §3.1) with a
/// single call:
///
/// ```dart
/// final confirmed = await AppDialog.confirm(
///   context,
///   title: 'Đăng xuất khỏi PandaDriver?',
///   message: 'Bạn sẽ cần đăng nhập lại để tiếp tục nhận chuyến.',
///   confirmLabel: 'Đăng xuất',
///   isDestructive: true,
/// );
/// ```
abstract final class AppDialog {
  static Future<bool> confirm(
    BuildContext context, {
    required String title,
    required String message,
    String confirmLabel = 'Xác nhận',
    String cancelLabel = 'Hủy',
    bool isDestructive = false,
  }) async {
    final result = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(title),
        content: Text(message),
        actionsPadding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          0,
          AppSpacing.lg,
          AppSpacing.lg,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: Text(cancelLabel),
          ),
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, true),
            style: TextButton.styleFrom(
              foregroundColor: isDestructive ? AppColors.error : AppColors.primary,
            ),
            child: Text(confirmLabel),
          ),
        ],
      ),
    );
    return result ?? false;
  }

  /// Single-action informational dialog (no cancel path) — used for
  /// placeholder/not-yet-implemented explainers where a snackbar would be
  /// missed (e.g. reached from a full-screen empty state rather than a
  /// visible tap target).
  static Future<void> info(
    BuildContext context, {
    required String title,
    required String message,
    String dismissLabel = 'Đã hiểu',
  }) {
    return showDialog<void>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: Text(title),
        content: Text(message),
        actionsPadding: const EdgeInsets.fromLTRB(
          AppSpacing.lg,
          0,
          AppSpacing.lg,
          AppSpacing.lg,
        ),
        actions: [
          SizedBox(
            width: double.infinity,
            child: AppButton.primary(
              label: dismissLabel,
              onPressed: () => Navigator.pop(dialogContext),
            ),
          ),
        ],
      ),
    );
  }
}
