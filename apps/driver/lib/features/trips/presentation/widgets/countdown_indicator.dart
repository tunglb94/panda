import 'package:flutter/material.dart';

/// Animated circular countdown, starting from [totalSeconds] (default 15).
/// Calls [onExpired] exactly once when it reaches zero — this is the only
/// thing that drives the "Expired" trip-offer transition, and it is purely
/// local/mock (no backend timeout is involved).
class CountdownIndicator extends StatefulWidget {
  const CountdownIndicator({
    super.key,
    this.totalSeconds = 15,
    required this.onExpired,
  });

  final int totalSeconds;
  final VoidCallback onExpired;

  @override
  State<CountdownIndicator> createState() => _CountdownIndicatorState();
}

class _CountdownIndicatorState extends State<CountdownIndicator>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: Duration(seconds: widget.totalSeconds),
    )..addStatusListener(_handleStatus);
    _controller.forward();
  }

  void _handleStatus(AnimationStatus status) {
    if (status == AnimationStatus.completed) {
      widget.onExpired();
    }
  }

  @override
  void dispose() {
    _controller.removeStatusListener(_handleStatus);
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, _) {
        final remaining = (widget.totalSeconds * (1 - _controller.value)).ceil();
        final urgent = remaining <= 5;
        return SizedBox(
          width: 64,
          height: 64,
          child: Stack(
            alignment: Alignment.center,
            children: [
              SizedBox(
                width: 64,
                height: 64,
                child: CircularProgressIndicator(
                  value: 1 - _controller.value,
                  strokeWidth: 5,
                  backgroundColor: Colors.grey.shade200,
                  valueColor: AlwaysStoppedAnimation(
                    urgent ? Colors.red.shade600 : primary,
                  ),
                ),
              ),
              Text(
                '$remaining',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  fontSize: 20,
                  color: urgent ? Colors.red.shade600 : Colors.black87,
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}
