import 'package:flutter/material.dart';

import '../../domain/models/rider_trip_status.dart';

/// Five-node connected stepper for the trip lifecycle — each node fills in
/// as the trip progresses, with a line connecting completed nodes.
///
/// Only the five "main flow" stages are shown here — payment sub-states
/// ([RiderTripStatus.paymentPending]/[paymentSuccess]/[settled]) and
/// [RiderTripStatus.cancelled] each have their own dedicated screen and
/// would otherwise overflow this row if all nine [RiderTripStatus] values
/// were rendered here.
class TripProgressIndicator extends StatelessWidget {
  const TripProgressIndicator({super.key, required this.status});

  final RiderTripStatus status;

  static const _stages = [
    RiderTripStatus.searchingDriver,
    RiderTripStatus.driverAssigned,
    RiderTripStatus.driverArriving,
    RiderTripStatus.inProgress,
    RiderTripStatus.completed,
  ];

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    final currentIndex = _stages.indexWhere((s) => s == status);

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: List.generate(_stages.length * 2 - 1, (i) {
        // Even indices are nodes, odd indices are connector lines.
        if (i.isOdd) {
          final leftStageIndex = i ~/ 2;
          final lineFilled = currentIndex >= 0 && leftStageIndex < currentIndex;
          return Expanded(
            child: Padding(
              padding: const EdgeInsets.only(top: 11),
              child: TweenAnimationBuilder<double>(
                tween: Tween(begin: 0, end: lineFilled ? 1 : 0),
                duration: const Duration(milliseconds: 400),
                curve: Curves.easeOut,
                builder: (context, t, _) => Container(
                  height: 3,
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(2),
                    gradient: LinearGradient(
                      colors: [primary, primary],
                      stops: [t, t],
                    ),
                    color: Colors.grey.shade200,
                  ),
                ),
              ),
            ),
          );
        }

        final stageIndex = i ~/ 2;
        final stage = _stages[stageIndex];
        final isReached = currentIndex >= 0 && stageIndex <= currentIndex;
        final isCurrent = stageIndex == currentIndex;

        return Column(
          children: [
            AnimatedContainer(
              duration: const Duration(milliseconds: 300),
              width: isCurrent ? 26 : 22,
              height: isCurrent ? 26 : 22,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: isReached ? primary : Colors.white,
                border: Border.all(
                  color: isReached ? primary : Colors.grey.shade300,
                  width: 2,
                ),
                boxShadow: isCurrent
                    ? [
                        BoxShadow(
                          color: primary.withValues(alpha: 0.35),
                          blurRadius: 8,
                          offset: const Offset(0, 2),
                        ),
                      ]
                    : null,
              ),
              child: isReached
                  ? Icon(
                      isCurrent ? Icons.circle : Icons.check,
                      size: isCurrent ? 10 : 13,
                      color: Colors.white,
                    )
                  : null,
            ),
            const SizedBox(height: 6),
            SizedBox(
              width: 54,
              child: Text(
                stage.shortLabel,
                textAlign: TextAlign.center,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  fontSize: 10,
                  fontWeight: isReached ? FontWeight.w700 : FontWeight.w500,
                  color: isReached ? primary : Colors.grey.shade400,
                ),
              ),
            ),
          ],
        );
      }),
    );
  }
}
