import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';

/// Small numeric/dot badge — unread notification counts, "99+" style
/// overflow. Pass [count] for a number, or omit it for a plain dot.
class AppBadge extends StatelessWidget {
  const AppBadge({super.key, this.count, this.color = AppColors.error});

  final int? count;
  final Color color;

  @override
  Widget build(BuildContext context) {
    final n = count;
    if (n != null && n <= 0) return const SizedBox.shrink();

    if (n == null) {
      return Container(
        width: 10,
        height: 10,
        decoration: BoxDecoration(color: color, shape: BoxShape.circle),
      );
    }

    return Container(
      constraints: const BoxConstraints(minWidth: 18, minHeight: 18),
      padding: const EdgeInsets.symmetric(horizontal: 4),
      decoration: BoxDecoration(color: color, shape: BoxShape.circle),
      alignment: Alignment.center,
      child: Text(
        n > 99 ? '99+' : '$n',
        style: const TextStyle(
          color: Colors.white,
          fontSize: 10,
          fontWeight: FontWeight.w700,
          height: 1,
        ),
      ),
    );
  }
}
