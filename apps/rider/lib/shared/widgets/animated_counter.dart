import 'package:flutter/material.dart';

/// Count-up animation for a formatted number (money totals, savings amounts).
/// Takes a formatter rather than a raw number so callers keep control of
/// currency symbols/decimal places while still getting the count-up motion.
/// Mirrors `apps/driver`'s `AnimatedCounter` exactly.
class AnimatedCounter extends StatelessWidget {
  const AnimatedCounter({
    super.key,
    required this.value,
    required this.format,
    this.style,
    this.duration = const Duration(milliseconds: 500),
  });

  final int value;
  final String Function(int) format;
  final TextStyle? style;
  final Duration duration;

  @override
  Widget build(BuildContext context) {
    return TweenAnimationBuilder<int>(
      tween: IntTween(begin: 0, end: value),
      duration: duration,
      curve: Curves.easeOut,
      builder: (context, v, _) => Text(format(v), style: style),
    );
  }
}
