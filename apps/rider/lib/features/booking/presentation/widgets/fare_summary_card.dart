import 'package:flutter/material.dart';

import '../../domain/models/mock_fare_calculator.dart';

/// Itemised fare preview. Mock data — see [MockFareBreakdown]. The total
/// animates with a short count-up tween whenever the vehicle or promo
/// selection changes upstream.
class FareSummaryCard extends StatelessWidget {
  const FareSummaryCard({super.key, required this.breakdown});

  final MockFareBreakdown breakdown;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Fare summary',
            style: Theme.of(context)
                .textTheme
                .titleSmall
                ?.copyWith(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 12),
          _FareRow(
              label: 'Base fare', value: breakdown.format(breakdown.baseFareCents)),
          _FareRow(
              label: 'Distance (est.)',
              value: breakdown.format(breakdown.distanceFareCents)),
          _FareRow(
              label: 'Time (est.)',
              value: breakdown.format(breakdown.timeFareCents)),
          _FareRow(
              label: 'Booking fee',
              value: breakdown.format(breakdown.bookingFeeCents)),
          if (breakdown.discountCents > 0)
            _FareRow(
              label: 'Promo discount',
              value: '-${breakdown.format(breakdown.discountCents)}',
              valueColor: Colors.green.shade700,
            ),
          const Divider(height: 20),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('Estimated total',
                  style: TextStyle(fontWeight: FontWeight.bold)),
              TweenAnimationBuilder<int>(
                tween: IntTween(begin: 0, end: breakdown.totalCents),
                duration: const Duration(milliseconds: 300),
                curve: Curves.easeOut,
                builder: (context, value, _) => Text(
                  breakdown.format(value),
                  style: const TextStyle(
                      fontWeight: FontWeight.bold, fontSize: 16),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _FareRow extends StatelessWidget {
  const _FareRow({required this.label, required this.value, this.valueColor});

  final String label;
  final String value;
  final Color? valueColor;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: TextStyle(color: Colors.grey.shade700)),
          Text(value,
              style: TextStyle(color: valueColor, fontWeight: FontWeight.w500)),
        ],
      ),
    );
  }
}
