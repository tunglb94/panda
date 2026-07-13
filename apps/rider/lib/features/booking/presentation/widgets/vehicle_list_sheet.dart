import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/shared/utils/currency_format.dart';
import 'package:rider/shared/widgets/app_bottom_sheet.dart';
import 'package:rider/shared/widgets/app_card.dart';

import '../../data/pricing_repository.dart';
import '../../domain/models/fare_estimate.dart';
import '../../domain/models/vehicle_option.dart';

enum _FilterTab { recommended, bike, car }

/// Opens the vehicle-list sheet (tabs + rows with capacity/price, styled
/// after Be/Xanh SM's vehicle picker). Returns the tapped [VehicleCategory],
/// or `null` if the rider dismissed the sheet without picking one.
///
/// Each row fetches its own real quote from the backend's Pricing service
/// (`PricingRepository.estimateFare`) — there is no client-side fare
/// computation anywhere in this sheet.
abstract final class VehicleListSheet {
  static Future<VehicleCategory?> show(
    BuildContext context, {
    required List<VehicleOption> options,
    required VehicleCategory selected,
    required LatLng pickup,
    required LatLng destination,
    required String tripType,
    required ApiClient apiClient,
  }) {
    return AppBottomSheet.show<VehicleCategory?>(
      context,
      title: 'Chọn loại xe',
      isScrollControlled: true,
      builder: (sheetContext) => _VehicleListBody(
        options: options,
        selected: selected,
        pickup: pickup,
        destination: destination,
        tripType: tripType,
        apiClient: apiClient,
      ),
    );
  }
}

class _VehicleListBody extends StatefulWidget {
  const _VehicleListBody({
    required this.options,
    required this.selected,
    required this.pickup,
    required this.destination,
    required this.tripType,
    required this.apiClient,
  });

  final List<VehicleOption> options;
  final VehicleCategory selected;
  final LatLng pickup;
  final LatLng destination;
  final String tripType;
  final ApiClient apiClient;

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
                return _VehicleRow(
                  key: ValueKey(option.category),
                  option: option,
                  pickup: widget.pickup,
                  destination: widget.destination,
                  tripType: widget.tripType,
                  apiClient: widget.apiClient,
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

/// One row — fetches its own real quote from the backend the moment it's
/// built (only visible rows get built, `ListView.separated` isn't lazy but
/// the sheet's row count is always small — 4 tiers today). No client-side
/// fare math; a failed quote shows "—", never a fabricated price.
class _VehicleRow extends StatefulWidget {
  const _VehicleRow({
    super.key,
    required this.option,
    required this.pickup,
    required this.destination,
    required this.tripType,
    required this.apiClient,
    required this.isSelected,
    required this.onTap,
  });

  final VehicleOption option;
  final LatLng pickup;
  final LatLng destination;
  final String tripType;
  final ApiClient apiClient;
  final bool isSelected;
  final VoidCallback? onTap;

  @override
  State<_VehicleRow> createState() => _VehicleRowState();
}

class _VehicleRowState extends State<_VehicleRow> {
  late final Future<FareEstimate>? _fareFuture = widget.option.isAvailable
      ? PricingRepository(widget.apiClient).estimateFare(
          pickup: widget.pickup,
          destination: widget.destination,
          serviceType: widget.option.category.backendKey,
          tripType: widget.tripType,
        )
      : null;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final option = widget.option;
    return Container(
      decoration: BoxDecoration(
        borderRadius: AppRadius.lgAll,
        border: Border.all(color: widget.isSelected ? AppColors.primary : Colors.transparent, width: 2),
      ),
      child: AppCard(
        animateIn: false,
        onTap: widget.onTap,
        padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
        color: widget.isSelected ? AppColors.primaryLight : AppColors.surface,
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
                child: option.imageAsset != null
                    ? Padding(
                        padding: const EdgeInsets.all(4),
                        child: Image.asset(option.imageAsset!, fit: BoxFit.contain),
                      )
                    : Icon(option.icon, color: AppColors.textPrimary, size: AppIconSize.lg),
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
              else
                FutureBuilder<FareEstimate>(
                  future: _fareFuture,
                  builder: (context, snapshot) {
                    if (!snapshot.hasData) {
                      return const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      );
                    }
                    if (snapshot.hasError) {
                      return Text('—', style: textTheme.bodyLarge?.copyWith(color: AppColors.textTertiary));
                    }
                    final fare = snapshot.data!;
                    return Text(
                      formatMoney(fare.total, fare.currencyCode),
                      style: textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w700, color: AppColors.primary),
                    );
                  },
                ),
            ],
          ),
        ),
      ),
    );
  }
}
