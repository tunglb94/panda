import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_icon_sizes.dart';
import '../../core/theme/app_spacing.dart';
import 'app_button.dart';
import 'mascot_image.dart';

/// Icon + title + subtitle + optional retry/action button — the one shape
/// for "there's nothing here" (empty list), "something went wrong" (error),
/// and "not built yet" (placeholder) states. Mirrors `apps/driver`'s
/// `AppEmptyState` exactly.
///
/// [mascotAsset] optionally replaces the icon-in-circle with a Panda mascot
/// (`shared/widgets/mascot_image.dart`) for the emotionally-significant
/// empty states called out in `docs/design/MASCOT_CATALOG.md` (no trips
/// yet, no vouchers, lost connection, GPS unavailable, no notifications) —
/// left unset, every other empty/error state keeps its plain icon exactly
/// as before, per the "don't stuff a mascot everywhere" design rule.
class AppEmptyState extends StatelessWidget {
  const AppEmptyState({
    super.key,
    required this.icon,
    required this.title,
    this.subtitle,
    this.actionLabel,
    this.onAction,
    this.iconColor,
    this.mascotAsset,
    this.mascotAnimation = MascotAnimation.scale,
  });

  /// Convenience constructor for an error state — red icon, "Thử lại" action.
  const AppEmptyState.error({
    super.key,
    this.title = 'Đã xảy ra lỗi',
    this.subtitle,
    this.onAction,
    this.actionLabel = 'Thử lại',
    this.mascotAsset,
    this.mascotAnimation = MascotAnimation.scale,
  })  : icon = Icons.error_outline,
        iconColor = AppColors.error;

  final IconData icon;
  final Color? iconColor;
  final String title;
  final String? subtitle;
  final String? actionLabel;
  final VoidCallback? onAction;

  /// File name under `assets/mascot/`, e.g. `mascot_no_connection.png`.
  final String? mascotAsset;
  final MascotAnimation mascotAnimation;

  @override
  Widget build(BuildContext context) {
    final color = iconColor ?? Theme.of(context).colorScheme.primary;
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 480),
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xxl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (mascotAsset != null)
                MascotImage(
                  asset: mascotAsset!,
                  size: MascotSize.large,
                  animation: mascotAnimation,
                )
              else
                Container(
                  padding: const EdgeInsets.all(AppSpacing.xl),
                  decoration: BoxDecoration(
                    color: color.withValues(alpha: 0.1),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(icon, size: AppIconSize.xxl, color: color),
                ),
              const SizedBox(height: AppSpacing.xl),
              Text(
                title,
                textAlign: TextAlign.center,
                style: Theme.of(context)
                    .textTheme
                    .titleMedium
                    ?.copyWith(fontWeight: FontWeight.w700),
              ),
              if (subtitle != null) ...[
                const SizedBox(height: AppSpacing.sm),
                Text(
                  subtitle!,
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
              if (actionLabel != null && onAction != null) ...[
                const SizedBox(height: AppSpacing.xl),
                AppButton.outline(
                  label: actionLabel!,
                  onPressed: onAction,
                  expand: false,
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
