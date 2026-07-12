import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_spacing.dart';

enum AppSnackbarType { info, success, warning, error }

/// Standard snackbar — icon + accent color per [AppSnackbarType]. Mirrors
/// `apps/driver`'s `AppSnackbar` exactly. Replaces bare
/// `ScaffoldMessenger.of(context).showSnackBar(const SnackBar(...))` calls.
abstract final class AppSnackbar {
  static void show(
    BuildContext context,
    String message, {
    AppSnackbarType type = AppSnackbarType.info,
  }) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          content: Row(
            children: [
              Icon(_iconFor(type), color: _accentFor(type), size: 20),
              const SizedBox(width: AppSpacing.sm + 2),
              Expanded(child: Text(message)),
            ],
          ),
        ),
      );
  }

  static void success(BuildContext context, String message) =>
      show(context, message, type: AppSnackbarType.success);

  static void warning(BuildContext context, String message) =>
      show(context, message, type: AppSnackbarType.warning);

  static void error(BuildContext context, String message) =>
      show(context, message, type: AppSnackbarType.error);

  static IconData _iconFor(AppSnackbarType type) => switch (type) {
        AppSnackbarType.success => Icons.check_circle,
        AppSnackbarType.warning => Icons.warning_amber_rounded,
        AppSnackbarType.error => Icons.error,
        AppSnackbarType.info => Icons.info,
      };

  static Color _accentFor(AppSnackbarType type) => switch (type) {
        AppSnackbarType.success => AppColors.primaryLight,
        AppSnackbarType.warning => AppColors.warning,
        AppSnackbarType.error => AppColors.errorLight,
        AppSnackbarType.info => AppColors.tint(Colors.white, 0.7),
      };
}
