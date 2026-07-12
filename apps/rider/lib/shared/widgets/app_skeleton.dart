import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_radius.dart';
import '../../core/theme/app_spacing.dart';

/// A single shimmering placeholder block, used while real content is still
/// loading. Mirrors `apps/driver`'s `AppSkeletonBox` exactly — the one
/// legitimate `AnimationController` in the shared widget set, since a
/// continuous back-and-forth sweep cannot be expressed as a one-shot
/// implicit animation. Disposed in `dispose()`.
class AppSkeletonBox extends StatefulWidget {
  const AppSkeletonBox({
    super.key,
    this.width,
    this.height = 14,
    this.borderRadius = AppRadius.smAll,
  });

  final double? width;
  final double height;
  final BorderRadius borderRadius;

  @override
  State<AppSkeletonBox> createState() => _AppSkeletonBoxState();
}

class _AppSkeletonBoxState extends State<AppSkeletonBox>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1100),
    )..repeat(reverse: true);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, _) {
        final t = _controller.value;
        return Container(
          width: widget.width,
          height: widget.height,
          decoration: BoxDecoration(
            color: Color.lerp(AppColors.divider, AppColors.border, t),
            borderRadius: widget.borderRadius,
          ),
        );
      },
    );
  }
}

/// A skeleton placeholder shaped like a list row — icon circle + two text
/// lines — used while the real list is loading instead of a bare spinner.
/// Mirrors `apps/driver`'s `AppSkeletonListTile`.
class AppSkeletonListTile extends StatelessWidget {
  const AppSkeletonListTile({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.md),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: AppRadius.lgAll,
        border: Border.all(color: AppColors.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              AppSkeletonBox(width: 90, height: 20, borderRadius: AppRadius.pillAll),
              const AppSkeletonBox(width: 60),
            ],
          ),
          const SizedBox(height: AppSpacing.sm),
          const AppSkeletonBox(width: double.infinity),
          const SizedBox(height: AppSpacing.xs),
          const AppSkeletonBox(width: 180),
        ],
      ),
    );
  }
}
