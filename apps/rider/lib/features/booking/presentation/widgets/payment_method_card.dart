import 'package:flutter/material.dart';

import '../../domain/models/mock_booking_catalog.dart';
import '../../domain/models/payment_method.dart';

/// Shows the selected payment method; tapping opens a mock method picker.
///
/// No Payment/Wallet backend exists yet (see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.1) — this never charges
/// anything, it only lets the user preview the picker UX.
class PaymentMethodCard extends StatelessWidget {
  const PaymentMethodCard({
    super.key,
    required this.selected,
    required this.onChanged,
  });

  final PaymentMethod selected;
  final ValueChanged<PaymentMethod> onChanged;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return InkWell(
      borderRadius: BorderRadius.circular(12),
      onTap: () => _showPicker(context),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.grey.shade200),
        ),
        child: Row(
          children: [
            Icon(selected.icon, color: primary),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(selected.label,
                      style: const TextStyle(fontWeight: FontWeight.w600)),
                  Text(
                    selected.subtitle,
                    style:
                        TextStyle(fontSize: 12, color: Colors.grey.shade500),
                  ),
                ],
              ),
            ),
            const Icon(Icons.keyboard_arrow_right, color: Colors.grey),
          ],
        ),
      ),
    );
  }

  Future<void> _showPicker(BuildContext context) async {
    final chosen = await showModalBottomSheet<PaymentMethod>(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (sheetContext) {
        final primary = Theme.of(sheetContext).colorScheme.primary;
        return SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Padding(
                padding: EdgeInsets.fromLTRB(20, 20, 20, 8),
                child: Align(
                  alignment: Alignment.centerLeft,
                  child: Text(
                    'Payment method',
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
                  ),
                ),
              ),
              ...MockBookingCatalog.paymentMethods.map((m) {
                final isSelected = m.type == selected.type;
                return ListTile(
                  leading: Icon(m.icon, color: isSelected ? primary : null),
                  title: Text(m.label),
                  subtitle: Text(m.subtitle),
                  trailing: AnimatedSwitcher(
                    duration: const Duration(milliseconds: 200),
                    child: isSelected
                        ? Icon(Icons.check_circle,
                            color: primary, key: const ValueKey('sel'))
                        : const SizedBox.shrink(key: ValueKey('unsel')),
                  ),
                  onTap: () => Navigator.pop(sheetContext, m),
                );
              }),
              const SizedBox(height: 12),
            ],
          ),
        );
      },
    );
    if (chosen != null) onChanged(chosen);
  }
}
