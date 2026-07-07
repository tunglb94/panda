import 'package:flutter/material.dart';

import '../../domain/models/rider_trip_status.dart';

/// Five-stage progress bar for the trip lifecycle, with a compact label row
/// beneath it. The fill animates smoothly whenever [status] changes.
class TripProgressIndicator extends StatelessWidget {
  const TripProgressIndicator({super.key, required this.status});

  final RiderTripStatus status;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        ClipRRect(
          borderRadius: BorderRadius.circular(4),
          child: TweenAnimationBuilder<double>(
            tween: Tween(begin: 0, end: status.progressValue),
            duration: const Duration(milliseconds: 500),
            curve: Curves.easeOut,
            builder: (context, value, _) => LinearProgressIndicator(
              value: value,
              minHeight: 6,
              backgroundColor: Colors.grey.shade200,
              valueColor: AlwaysStoppedAnimation(primary),
            ),
          ),
        ),
        const SizedBox(height: 8),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: RiderTripStatus.values.map((s) {
            final reached = s.progressValue <= status.progressValue;
            return Text(
              s.shortLabel,
              style: TextStyle(
                fontSize: 10,
                fontWeight: reached ? FontWeight.w700 : FontWeight.w400,
                color: reached ? primary : Colors.grey.shade400,
              ),
            );
          }).toList(),
        ),
      ],
    );
  }
}
