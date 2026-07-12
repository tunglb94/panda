import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_bottom_sheet.dart';
import 'package:rider/shared/widgets/app_card.dart';

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
    return AppCard(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg, vertical: AppSpacing.md),
      onTap: () => _showPicker(context),
      child: Row(
        children: [
          Icon(selected.icon, color: AppColors.primary),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(selected.label, style: Theme.of(context).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600)),
                Text(selected.subtitle, style: Theme.of(context).textTheme.labelMedium),
              ],
            ),
          ),
          const Icon(Icons.keyboard_arrow_right, color: AppColors.textTertiary),
        ],
      ),
    );
  }

  Future<void> _showPicker(BuildContext context) async {
    final chosen = await AppBottomSheet.show<PaymentMethod>(
      context,
      title: 'Phương thức thanh toán',
      builder: (sheetContext) => Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          ...MockBookingCatalog.paymentMethods.map((m) {
            final isSelected = m.type == selected.type;
            return ListTile(
              leading: Icon(m.icon, color: isSelected ? AppColors.primary : null),
              title: Text(m.label),
              subtitle: Text(m.subtitle),
              trailing: AnimatedSwitcher(
                duration: const Duration(milliseconds: 200),
                child: isSelected
                    ? const Icon(Icons.check_circle, color: AppColors.primary, key: ValueKey('sel'))
                    : const SizedBox.shrink(key: ValueKey('unsel')),
              ),
              onTap: () => Navigator.pop(sheetContext, m),
            );
          }),
        ],
      ),
    );
    if (chosen != null) onChanged(chosen);
  }
}
