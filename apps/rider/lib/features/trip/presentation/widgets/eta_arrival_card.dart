import 'package:flutter/material.dart';

/// Shows the mock ETA and estimated arrival time side by side.
class EtaArrivalCard extends StatelessWidget {
  const EtaArrivalCard({super.key, required this.eta, required this.arrivalLabel});

  final Duration eta;
  final String arrivalLabel;

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
            child: _EtaColumn(label: 'ETA', value: '${eta.inMinutes} min'),
          ),
          Container(width: 1, height: 32, color: Colors.grey.shade300),
          Expanded(
            child: _EtaColumn(
              label: 'Estimated arrival',
              value: arrivalLabel,
              alignEnd: true,
            ),
          ),
        ],
      ),
    );
  }
}

class _EtaColumn extends StatelessWidget {
  const _EtaColumn({
    required this.label,
    required this.value,
    this.alignEnd = false,
  });

  final String label;
  final String value;
  final bool alignEnd;

  @override
  Widget build(BuildContext context) {
    final crossAxis =
        alignEnd ? CrossAxisAlignment.end : CrossAxisAlignment.start;
    return Column(
      crossAxisAlignment: crossAxis,
      children: [
        Text(label, style: TextStyle(fontSize: 11, color: Colors.grey.shade500)),
        const SizedBox(height: 2),
        Text(value, style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 15)),
      ],
    );
  }
}
