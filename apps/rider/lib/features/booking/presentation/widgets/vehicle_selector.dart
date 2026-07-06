import 'package:flutter/material.dart';

import '../../domain/models/mock_fare_calculator.dart';
import '../../domain/models/vehicle_option.dart';

/// Horizontally scrollable list of ride options (Economy tier only, per
/// DOC-0002 §6.19: car, motorcycle, van). Selecting a tile updates the Fare
/// Summary above/below it via [onSelected]. Mock data only.
class VehicleSelector extends StatelessWidget {
  const VehicleSelector({
    super.key,
    required this.options,
    required this.selected,
    required this.distanceKm,
    required this.durationMin,
    required this.onSelected,
  });

  final List<VehicleOption> options;
  final VehicleCategory selected;
  final double distanceKm;
  final double durationMin;
  final ValueChanged<VehicleCategory> onSelected;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 108,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        itemCount: options.length,
        separatorBuilder: (_, _) => const SizedBox(width: 10),
        itemBuilder: (context, index) {
          final option = options[index];
          final isSelected = option.category == selected;
          final fare = MockFareBreakdown.calculate(
            vehicle: option,
            distanceKm: distanceKm,
            durationMin: durationMin,
          );
          return _VehicleTile(
            option: option,
            isSelected: isSelected,
            estimatedTotal: fare.format(fare.totalCents),
            onTap: () => onSelected(option.category),
          );
        },
      ),
    );
  }
}

class _VehicleTile extends StatelessWidget {
  const _VehicleTile({
    required this.option,
    required this.isSelected,
    required this.estimatedTotal,
    required this.onTap,
  });

  final VehicleOption option;
  final bool isSelected;
  final String estimatedTotal;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 220),
        curve: Curves.easeOut,
        width: 108,
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 8),
        decoration: BoxDecoration(
          color: isSelected ? primary.withValues(alpha: 0.08) : Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(
            color: isSelected ? primary : Colors.grey.shade200,
            width: isSelected ? 1.6 : 1,
          ),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            AnimatedScale(
              scale: isSelected ? 1.1 : 1.0,
              duration: const Duration(milliseconds: 220),
              curve: Curves.easeOut,
              child: Icon(option.icon, color: primary, size: 26),
            ),
            const SizedBox(height: 6),
            Text(
              option.label,
              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
            ),
            const SizedBox(height: 2),
            Text(
              estimatedTotal,
              style: TextStyle(
                fontSize: 12,
                color: Colors.grey.shade600,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
