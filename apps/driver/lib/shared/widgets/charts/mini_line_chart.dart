import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';

/// A minimal sparkline — `CustomPainter`, no charting package. Draws a flat
/// mid-line when there's no real data instead of a fabricated trend, so an
/// all-zero series still reads honestly as "nothing to show yet".
class MiniLineChart extends StatelessWidget {
  const MiniLineChart({
    super.key,
    required this.values,
    this.height = 48,
    this.color = AppColors.primary,
  });

  final List<int> values;
  final double height;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return TweenAnimationBuilder<double>(
      tween: Tween(begin: 0, end: 1),
      duration: const Duration(milliseconds: 600),
      curve: Curves.easeOut,
      builder: (context, t, _) => SizedBox(
        height: height,
        width: double.infinity,
        child: CustomPaint(
          painter: _LinePainter(values: values, color: color, progress: t),
        ),
      ),
    );
  }
}

class _LinePainter extends CustomPainter {
  _LinePainter({required this.values, required this.color, required this.progress});

  final List<int> values;
  final Color color;
  final double progress;

  @override
  void paint(Canvas canvas, Size size) {
    if (values.isEmpty) return;
    final maxValue = values.reduce((a, b) => a > b ? a : b);
    final hasData = maxValue > 0;

    final points = <Offset>[];
    for (var i = 0; i < values.length; i++) {
      final x = values.length == 1 ? 0.0 : size.width * i / (values.length - 1);
      final fraction = hasData ? values[i] / maxValue : 0.5;
      final y = hasData ? size.height * (1 - fraction) : size.height / 2;
      points.add(Offset(x, y));
    }

    final linePaint = Paint()
      ..color = hasData ? color : AppColors.border
      ..strokeWidth = 2.5
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round;

    final path = Path()..moveTo(points.first.dx, points.first.dy);
    for (final p in points.skip(1)) {
      path.lineTo(p.dx, p.dy);
    }

    // Animate the line drawing in left-to-right via a clip.
    canvas.save();
    canvas.clipRect(Rect.fromLTWH(0, 0, size.width * progress, size.height));
    canvas.drawPath(path, linePaint);
    canvas.restore();

    if (hasData) {
      final dotPaint = Paint()..color = color;
      canvas.drawCircle(points.last, 4 * progress, dotPaint);
    }
  }

  @override
  bool shouldRepaint(covariant _LinePainter oldDelegate) =>
      oldDelegate.values != values || oldDelegate.progress != progress;
}
