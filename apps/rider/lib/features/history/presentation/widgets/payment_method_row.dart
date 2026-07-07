import 'package:flutter/material.dart';

import 'package:rider/features/booking/domain/models/payment_method.dart';

/// Read-only payment method display for a past trip.
///
/// Deliberately not the interactive `PaymentMethodCard` (Booking module) —
/// that widget opens a picker on tap, which is the wrong affordance for a
/// trip that already happened and can't have its payment method changed.
/// Reuses the `PaymentMethod` model itself, just not that widget.
class PaymentMethodRow extends StatelessWidget {
  const PaymentMethodRow({super.key, required this.method});

  final PaymentMethod method;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        children: [
          Icon(method.icon, color: primary),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(method.label, style: const TextStyle(fontWeight: FontWeight.w600)),
                Text(
                  method.subtitle,
                  style: TextStyle(fontSize: 12, color: Colors.grey.shade500),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
