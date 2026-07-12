import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';

/// A circular progress ring with a centered label — used for Driver Level
/// progress and any "X of Y" style stat. When [value] is null (no real
/// figure available), renders a neutral gray ring at a fixed low fraction
/// instead of a fabricated percentage, with [placeholderLabel] ("—" by
/// default) in the center rather than a made-up number.
class ProgressRing extends StatelessWidget {
  const ProgressRing({
    super.key,
    required this.value,
    required this.centerLabel,
    this.size = 64,
    this.strokeWidth = 6,
    this.color = AppColors.primary,
    this.placeholderLabel = '—',
  });

  final double? value; // 0..1, null = no real data
  final String centerLabel;
  final double size;
  final double strokeWidth;
  final Color color;
  final String placeholderLabel;

  @override
  Widget build(BuildContext context) {
    final hasData = value != null;
    return SizedBox(
      width: size,
      height: size,
      child: Stack(
        alignment: Alignment.center,
        children: [
          TweenAnimationBuilder<double>(
            tween: Tween(begin: 0, end: hasData ? value!.clamp(0, 1) : 0.08),
            duration: const Duration(milliseconds: 600),
            curve: Curves.easeOut,
            builder: (context, t, _) => CircularProgressIndicator(
              value: t,
              strokeWidth: strokeWidth,
              backgroundColor: AppColors.divider,
              valueColor: AlwaysStoppedAnimation(hasData ? color : AppColors.textTertiary),
            ),
          ),
          Text(
            hasData ? centerLabel : placeholderLabel,
            style: TextStyle(
              fontSize: size * 0.22,
              fontWeight: FontWeight.w700,
              color: hasData ? AppColors.textPrimary : AppColors.textTertiary,
            ),
          ),
        ],
      ),
    );
  }
}
