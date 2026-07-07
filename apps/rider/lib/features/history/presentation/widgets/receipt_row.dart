import 'package:flutter/material.dart';

/// Label/value row styled for the Receipt screen (monospace-leaning, no
/// card chrome — distinct from the Booking module's `FareSummaryCard` so a
/// receipt actually reads like a receipt).
class ReceiptRow extends StatelessWidget {
  const ReceiptRow({
    super.key,
    required this.label,
    required this.value,
    this.bold = false,
  });

  final String label;
  final String value;
  final bool bold;

  @override
  Widget build(BuildContext context) {
    final style = TextStyle(
      fontFamily: 'monospace',
      fontWeight: bold ? FontWeight.bold : FontWeight.normal,
      fontSize: bold ? 15 : 13,
    );
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: style),
          const SizedBox(width: 12),
          Expanded(
            child: Text(value, style: style, textAlign: TextAlign.right),
          ),
        ],
      ),
    );
  }
}

/// Dashed horizontal rule, the classic receipt-tear-off look.
class ReceiptDivider extends StatelessWidget {
  const ReceiptDivider({super.key});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 12,
      child: LayoutBuilder(
        builder: (context, constraints) {
          const dashWidth = 5.0;
          const dashSpace = 4.0;
          final count = (constraints.maxWidth / (dashWidth + dashSpace)).floor();
          return Row(
            children: List.generate(
              count,
              (_) => Padding(
                padding: const EdgeInsets.only(right: dashSpace),
                child: Container(width: dashWidth, height: 1.4, color: Colors.grey.shade400),
              ),
            ),
          );
        },
      ),
    );
  }
}
