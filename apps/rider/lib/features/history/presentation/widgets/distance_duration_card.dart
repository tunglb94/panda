import 'package:flutter/material.dart';

/// Distance + duration stat row for the Trip Detail screen. Styled the same
/// way as `EtaArrivalCard` (Trip module) and `ProfileStatsRow` (Profile
/// module) for visual consistency across features.
class DistanceDurationCard extends StatelessWidget {
  const DistanceDurationCard({
    super.key,
    required this.distanceKm,
    required this.durationMin,
  });

  final double distanceKm;
  final double durationMin;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        children: [
          Expanded(
            child: _StatColumn(
              label: 'Distance',
              value: '${distanceKm.toStringAsFixed(1)} km',
            ),
          ),
          Container(width: 1, height: 32, color: Colors.grey.shade300),
          Expanded(
            child: _StatColumn(
              label: 'Duration',
              value: '${durationMin.round()} min',
            ),
          ),
        ],
      ),
    );
  }
}

class _StatColumn extends StatelessWidget {
  const _StatColumn({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: TextStyle(fontSize: 11, color: Colors.grey.shade500)),
        const SizedBox(height: 2),
        Text(value, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 15)),
      ],
    );
  }
}
