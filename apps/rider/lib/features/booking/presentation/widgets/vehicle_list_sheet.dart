import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/widgets/app_bottom_sheet.dart';
import 'package:rider/shared/widgets/app_card.dart';

import '../../domain/models/mock_fare_calculator.dart';
import '../../domain/models/vehicle_option.dart';

enum _FilterTab { recommended, bike, car }

/// Opens the vehicle-list sheet (tabs + rows with capacity/price, styled
/// after Be/Xanh SM's vehicle picker). Returns the tapped [VehicleCategory],
/// or `null` if the rider dismissed the sheet without picking one.
abstract final class VehicleListSheet {
  static Future<VehicleCategory?> show(
    BuildContext context, {
    required List<VehicleOption> options,
    required VehicleCategory selected,
    required double distanceKm,
    required double durationMin,
  }) {
    return AppBottomSheet.show<VehicleCategory?>(
      context,
      title: 'Chọn loại xe',
      isScrollControlled: true,
      builder: (sheetContext) => _VehicleListBody(
        options: options,
        selected: selected,
        distanceKm: distanceKm,
        durationMin: durationMin,
      ),
    );
  }
}

class _VehicleListBody extends StatefulWidget {
  const _VehicleListBody({
    required this.options,
    required this.selected,
    required this.distanceKm,
    required this.durationMin,
  });

  final List<VehicleOption> options;
  final VehicleCategory selected;
  final double distanceKm;
  final double durationMin;

  @override
  State<_VehicleListBody> createState() => _VehicleListBodyState();
}

class _VehicleListBodyState extends State<_VehicleListBody> {
  _FilterTab _tab = _FilterTab.recommended;

  @override
  Widget build(BuildContext context) {
    final visible = widget.options.where((o) {
      switch (_tab) {
        case _FilterTab.recommended:
          return true;
        case _FilterTab.bike:
          return o.category.family == VehicleFamily.bike;
        case _FilterTab.car:
          return o.category.family == VehicleFamily.car;
      }
    }).toList();

    return ConstrainedBox(
      constraints: BoxConstraints(maxHeight: MediaQuery.of(context).size.height * 0.75),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          _TabBar(selected: _tab, onChanged: (t) => setState(() => _tab = t)),
          const SizedBox(height: AppSpacing.md),
          Flexible(
            child: ListView.separated(
              shrinkWrap: true,
              itemCount: visible.length,
              separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
              itemBuilder: (context, i) {
                final option = visible[i];
                final isSelected = option.category == widget.selected;
                final fare = option.isAvailable
                    ? MockFareBreakdown.calculate(
                        vehicle: option,
                        distanceKm: widget.distanceKm,
                        durationMin: widget.durationMin,
                      )
                    : null;
                return _VehicleRow(
                  option: option,
                  fare: fare,
                  isSelected: isSelected,
                  onTap: option.isAvailable ? () => Navigator.of(context).pop(option.category) : null,
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _TabBar extends StatelessWidget {
  const _TabBar({required this.selected, required this.onChanged});

  final _FilterTab selected;
  final ValueChanged<_FilterTab> onChanged;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        _TabChip(label: 'Đề xuất', selected: selected == _FilterTab.recommended, onTap: () => onChanged(_FilterTab.recommended)),
        const SizedBox(width: AppSpacing.sm),
        _TabChip(label: 'Xe máy', selected: selected == _FilterTab.bike, onTap: () => onChanged(_FilterTab.bike)),
        const SizedBox(width: AppSpacing.sm),
        _TabChip(label: 'Ô tô', selected: selected == _FilterTab.car, onTap: () => onChanged(_FilterTab.car)),
      ],
    );
  }
}

class _TabChip extends StatelessWidget {
  const _TabChip({required this.label, required this.selected, required this.onTap});

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: AppRadius.pillAll,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 180),
        padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
        decoration: BoxDecoration(
          borderRadius: AppRadius.pillAll,
          color: selected ? AppColors.primaryLight : AppColors.surfaceAlt,
          border: Border.all(color: selected ? AppColors.primary : Colors.transparent, width: 1.5),
        ),
        child: Text(
          label,
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: selected ? AppColors.primaryDark : AppColors.textSecondary,
                fontWeight: selected ? FontWeight.w700 : FontWeight.w500,
              ),
        ),
      ),
    );
  }
}

class _VehicleRow extends StatelessWidget {
  const _VehicleRow({
    required this.option,
    required this.fare,
    required this.isSelected,
    required this.onTap,
  });

  final VehicleOption option;
  final MockFareBreakdown? fare;
  final bool isSelected;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    return Container(
      decoration: BoxDecoration(
        borderRadius: AppRadius.lgAll,
        border: Border.all(color: isSelected ? AppColors.primary : Colors.transparent, width: 2),
      ),
      child: AppCard(
        animateIn: false,
        onTap: onTap,
        padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
        color: isSelected ? AppColors.primaryLight : AppColors.surface,
        child: Opacity(
          opacity: option.isAvailable ? 1 : 0.5,
          child: Row(
            children: [
              Container(
                width: 44,
                height: 44,
                decoration: BoxDecoration(
                  color: AppColors.surfaceAlt,
                  borderRadius: AppRadius.mdAll,
                ),
                child: Icon(option.icon, color: AppColors.textPrimary, size: AppIconSize.lg),
              ),
              const SizedBox(width: AppSpacing.md),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(option.label, style: textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w700)),
                    const SizedBox(height: 2),
                    Row(
                      children: [
                        Icon(Icons.person_outline, size: AppIconSize.sm, color: AppColors.textTertiary),
                        const SizedBox(width: 2),
                        Text('${option.capacity}', style: textTheme.bodySmall?.copyWith(color: AppColors.textTertiary)),
                      ],
                    ),
                  ],
                ),
              ),
              if (!option.isAvailable)
                Text('Chưa khả dụng', style: textTheme.bodySmall?.copyWith(color: AppColors.textTertiary))
              else if (fare != null)
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    if (option.originalPriceCents != null)
                      Text(
                        fare!.format(option.originalPriceCents!),
                        style: textTheme.bodySmall?.copyWith(
                          color: AppColors.textTertiary,
                          decoration: TextDecoration.lineThrough,
                        ),
                      ),
                    Text(
                      fare!.format(fare!.totalCents),
                      style: textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w700, color: AppColors.primary),
                    ),
                  ],
                ),
            ],
          ),
        ),
      ),
    );
  }
}
