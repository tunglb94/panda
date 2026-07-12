import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_radius.dart';

/// A minimal bar chart — no charting package, just a `Row` of animated
/// bars. If every value is 0 (no real earnings yet), draws flat nub bars
/// instead of a misleading empty canvas, so "no data" reads as an honest
/// empty state rather than a broken chart.
class MiniBarChart extends StatelessWidget {
  const MiniBarChart({
    super.key,
    required this.values,
    required this.labels,
    this.height = 80,
    this.barColor = AppColors.primary,
  });

  final List<int> values;
  final List<String> labels;
  final double height;
  final Color barColor;

  @override
  Widget build(BuildContext context) {
    final maxValue = values.isEmpty ? 0 : values.reduce((a, b) => a > b ? a : b);
    final hasData = maxValue > 0;

    return SizedBox(
      height: height + 20,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: List.generate(values.length, (i) {
          final fraction = hasData ? values[i] / maxValue : 0.0;
          final isToday = i == values.length - 1;
          return Expanded(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 3),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  SizedBox(
                    height: height,
                    child: Align(
                      alignment: Alignment.bottomCenter,
                      child: TweenAnimationBuilder<double>(
                        tween: Tween(begin: 0, end: fraction),
                        duration: const Duration(milliseconds: 500),
                        curve: Curves.easeOut,
                        builder: (context, t, _) => Container(
                          height: hasData ? (height * t).clamp(4, height) : 4,
                          decoration: BoxDecoration(
                            color: isToday
                                ? barColor
                                : barColor.withValues(alpha: 0.25),
                            borderRadius: AppRadius.smAll,
                          ),
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    labels[i],
                    style: TextStyle(
                      fontSize: 10,
                      fontWeight: isToday ? FontWeight.w700 : FontWeight.w500,
                      color: isToday ? AppColors.textPrimary : AppColors.textTertiary,
                    ),
                  ),
                ],
              ),
            ),
          );
        }),
      ),
    );
  }
}
