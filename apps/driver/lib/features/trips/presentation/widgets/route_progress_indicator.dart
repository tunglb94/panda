import 'package:flutter/material.dart';

import '../../domain/models/route_progress_model.dart';

/// Mock progress bar for the pickup leg — no map, no GPS. Shows [progress]
/// (100 → 0) as a linear bar plus a color-coded traffic badge; the trailing
/// label reads "Arrived" instead of "0% remaining" once [progress] hits 0.
class RouteProgressIndicator extends StatelessWidget {
  const RouteProgressIndicator({super.key, required this.progress, required this.trafficLevel});

  final int progress;
  final TrafficLevel trafficLevel;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'Route progress',
              style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
            ),
            _TrafficBadge(trafficLevel: trafficLevel),
          ],
        ),
        const SizedBox(height: 8),
        TweenAnimationBuilder<double>(
          tween: Tween(begin: 0, end: progress / 100),
          duration: const Duration(milliseconds: 400),
          builder: (context, value, _) => ClipRRect(
            borderRadius: BorderRadius.circular(6),
            child: LinearProgressIndicator(
              value: value,
              minHeight: 8,
              backgroundColor: Colors.grey.shade200,
              color: primary,
            ),
          ),
        ),
        const SizedBox(height: 6),
        Text(
          progress <= 0 ? 'Arrived' : '$progress% remaining',
          style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
        ),
      ],
    );
  }
}

class _TrafficBadge extends StatelessWidget {
  const _TrafficBadge({required this.trafficLevel});

  final TrafficLevel trafficLevel;

  @override
  Widget build(BuildContext context) {
    final color = switch (trafficLevel) {
      TrafficLevel.normal => Colors.green.shade600,
      TrafficLevel.slow => Colors.orange.shade700,
      TrafficLevel.heavy => Colors.red.shade600,
    };
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(color: color.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(8)),
      child: Text(
        trafficLevel.label,
        style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: color),
      ),
    );
  }
}
