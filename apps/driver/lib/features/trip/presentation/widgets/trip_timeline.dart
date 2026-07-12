import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';

/// The 4 stages of an accepted trip, rendered as a connected-dot stepper —
/// mirrors `apps/rider`'s `TripProgressIndicator` visual language so the two
/// apps read as one product family.
enum TripTimelineStage { accepted, pickup, inTrip, done }

class TripTimeline extends StatelessWidget {
  const TripTimeline({super.key, required this.current});

  final TripTimelineStage current;

  static const _stages = TripTimelineStage.values;
  static const _labels = ['Đã nhận', 'Đến đón', 'Đang chạy', 'Hoàn tất'];

  @override
  Widget build(BuildContext context) {
    final currentIndex = _stages.indexOf(current);
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: List.generate(_stages.length * 2 - 1, (i) {
        if (i.isOdd) {
          final leftIndex = i ~/ 2;
          final filled = leftIndex < currentIndex;
          return Expanded(
            child: Padding(
              padding: const EdgeInsets.only(top: 9),
              child: Container(
                height: 3,
                decoration: BoxDecoration(
                  color: filled ? AppColors.primary : AppColors.border,
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
          );
        }
        final stageIndex = i ~/ 2;
        final reached = stageIndex <= currentIndex;
        final isCurrent = stageIndex == currentIndex;
        return Column(
          children: [
            AnimatedContainer(
              duration: const Duration(milliseconds: 250),
              width: isCurrent ? 22 : 18,
              height: isCurrent ? 22 : 18,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: reached ? AppColors.primary : Colors.white,
                border: Border.all(
                  color: reached ? AppColors.primary : AppColors.border,
                  width: 2,
                ),
              ),
              child: reached
                  ? Icon(
                      isCurrent ? Icons.circle : Icons.check,
                      size: isCurrent ? 8 : 11,
                      color: Colors.white,
                    )
                  : null,
            ),
            const SizedBox(height: 4),
            SizedBox(
              width: 60,
              child: Text(
                _labels[stageIndex],
                textAlign: TextAlign.center,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  fontSize: 10,
                  fontWeight: reached ? FontWeight.w700 : FontWeight.w500,
                  color: reached ? AppColors.primary : AppColors.textTertiary,
                ),
              ),
            ),
          ],
        );
      }),
    );
  }
}
