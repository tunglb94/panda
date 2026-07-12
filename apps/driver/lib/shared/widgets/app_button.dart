import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_radius.dart';
import '../../core/theme/app_spacing.dart';
import 'pressable_scale.dart';

enum AppButtonVariant { primary, outline, danger, text }

/// The one button widget for PandaDriver. Every primary/secondary/danger
/// action in the app should go through this rather than reaching for
/// `FilledButton`/`OutlinedButton` directly with ad hoc `style:` overrides —
/// centralizing here is what makes the in-flight loading spinner and the
/// press-down scale feedback consistent everywhere for free.
///
/// Every button gets, for free:
/// - a press-down scale (via [PressScaleObserver] — safe to wrap because it
///   observes raw pointer events rather than competing for the tap gesture)
/// - the platform ripple (unchanged, from the underlying Material button)
/// - a visually distinct disabled state (see `AppTheme`'s button themes)
/// - a loading morph (spinner replaces the label, [isLoading])
/// - an optional success morph (checkmark replaces the label briefly,
///   [isSuccess] — caller-controlled, matching the same external-flag
///   pattern as [isLoading] rather than an internal timer, so the caller
///   decides how long "success" stays visible)
class AppButton extends StatelessWidget {
  const AppButton({
    super.key,
    required this.label,
    required this.onPressed,
    this.variant = AppButtonVariant.primary,
    this.icon,
    this.isLoading = false,
    this.isSuccess = false,
    this.expand = true,
  });

  const AppButton.primary({
    super.key,
    required this.label,
    required this.onPressed,
    this.icon,
    this.isLoading = false,
    this.isSuccess = false,
    this.expand = true,
  }) : variant = AppButtonVariant.primary;

  const AppButton.outline({
    super.key,
    required this.label,
    required this.onPressed,
    this.icon,
    this.isLoading = false,
    this.isSuccess = false,
    this.expand = true,
  }) : variant = AppButtonVariant.outline;

  const AppButton.danger({
    super.key,
    required this.label,
    required this.onPressed,
    this.icon,
    this.isLoading = false,
    this.isSuccess = false,
    this.expand = true,
  }) : variant = AppButtonVariant.danger;

  const AppButton.text({
    super.key,
    required this.label,
    required this.onPressed,
    this.icon,
    this.isLoading = false,
    this.isSuccess = false,
    this.expand = false,
  }) : variant = AppButtonVariant.text;

  final String label;
  final VoidCallback? onPressed;
  final AppButtonVariant variant;
  final IconData? icon;
  final bool isLoading;

  /// Shows a brief checkmark-and-label morph instead of the normal label.
  /// Intended for the handful of genuinely celebratory confirmations in the
  /// app (accepting an offer, submitting a rating) — not every button needs
  /// this, so it's opt-in per call site rather than automatic.
  final bool isSuccess;

  /// Whether the button fills the width available to it (the default for
  /// every primary trip-flow action) or sizes to its content (typical for
  /// inline/text buttons).
  final bool expand;

  @override
  Widget build(BuildContext context) {
    final effectiveOnPressed = (isLoading || isSuccess) ? null : onPressed;
    final child = AnimatedSwitcher(
      duration: const Duration(milliseconds: 180),
      transitionBuilder: (child, animation) =>
          ScaleTransition(scale: animation, child: child),
      child: switch ((isLoading, isSuccess)) {
        (true, _) => SizedBox(
            key: const ValueKey('loading'),
            height: 20,
            width: 20,
            child: CircularProgressIndicator(
              strokeWidth: 2.4,
              color: _spinnerColor(context),
            ),
          ),
        (false, true) => Row(
            key: const ValueKey('success'),
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.check_circle, size: 20),
              const SizedBox(width: AppSpacing.sm),
              Flexible(
                child: Text(label, overflow: TextOverflow.ellipsis, maxLines: 1),
              ),
            ],
          ),
        (false, false) => _LabelRow(key: const ValueKey('label'), label: label, icon: icon),
      },
    );

    final button = switch (variant) {
      AppButtonVariant.primary => FilledButton(onPressed: effectiveOnPressed, child: child),
      AppButtonVariant.outline => OutlinedButton(onPressed: effectiveOnPressed, child: child),
      AppButtonVariant.text => TextButton(onPressed: effectiveOnPressed, child: child),
      AppButtonVariant.danger => OutlinedButton(
          onPressed: effectiveOnPressed,
          style: OutlinedButton.styleFrom(
            foregroundColor: AppColors.error,
            disabledForegroundColor: AppColors.textTertiary,
            side: const BorderSide(color: AppColors.error),
            minimumSize: const Size.fromHeight(52),
            shape: const RoundedRectangleBorder(borderRadius: AppRadius.mdAll),
          ).copyWith(
            side: WidgetStateProperty.resolveWith((states) {
              if (states.contains(WidgetState.disabled)) {
                return const BorderSide(color: AppColors.border);
              }
              return const BorderSide(color: AppColors.error);
            }),
          ),
          child: child,
        ),
    };

    final sized = expand ? SizedBox(width: double.infinity, child: button) : button;
    return PressScaleObserver(scale: 0.97, child: sized);
  }

  Color _spinnerColor(BuildContext context) => switch (variant) {
        AppButtonVariant.primary => AppColors.textOnPrimary,
        AppButtonVariant.outline => AppColors.primary,
        AppButtonVariant.danger => AppColors.error,
        AppButtonVariant.text => AppColors.primary,
      };
}

class _LabelRow extends StatelessWidget {
  const _LabelRow({super.key, required this.label, this.icon});

  final String label;
  final IconData? icon;

  @override
  Widget build(BuildContext context) {
    if (icon == null) {
      return Text(label, overflow: TextOverflow.ellipsis, maxLines: 1);
    }
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 20),
        const SizedBox(width: AppSpacing.sm),
        Flexible(
          child: Text(label, overflow: TextOverflow.ellipsis, maxLines: 1),
        ),
      ],
    );
  }
}
