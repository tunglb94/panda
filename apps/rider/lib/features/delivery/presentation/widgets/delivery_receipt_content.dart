import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/booking/presentation/widgets/trip_point_cards.dart';
import 'package:rider/features/delivery/domain/models/delivery_status.dart';
import 'package:rider/features/trip/data/trip_repository.dart';
import 'package:rider/shared/utils/currency_format.dart';
import 'package:rider/shared/widgets/app_card.dart';
import 'package:rider/shared/widgets/app_status_chip.dart';

/// Delivery-specific receipt — a distinct widget from `TripReceiptContent`,
/// per the production-pass requirement that Delivery not reuse Ride's
/// receipt. Every row is either a real field already on `TripDetail`
/// (pickup/dropoff address, delivery status) or an explicit "Chưa cập nhật"
/// placeholder — the underlying Trip's `final_fare` never settles for a
/// delivery trip today (`CompleteDeliveryUseCase` deliberately leaves
/// `Trip.Status` at `in_progress`, since Pricing was never extended for
/// Delivery fares — see the Delivery wire-contract audit), so the total
/// shown here is honestly labelled as the pre-booking estimate, never as a
/// paid/settled amount.
class DeliveryReceiptContent extends StatelessWidget {
  const DeliveryReceiptContent({
    super.key,
    required this.trip,
    this.receiverName,
    this.receiverPhone,
    this.estimatedFareCents,
    this.currency,
  });

  final TripDetail trip;
  final String? receiverName;
  final String? receiverPhone;

  /// Captured client-side at booking time (see `DeliveryFormPage`'s live
  /// estimate) — not re-derived here, since there is no real backend fare
  /// for delivery trips to fetch.
  final int? estimatedFareCents;
  final String? currency;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final status = DeliveryStatus.fromWire(trip.deliveryStatus);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Row(
          children: [
            AppStatusChip(label: status.label, color: status.isTerminal ? AppColors.primary : AppColors.warning),
          ],
        ),
        const SizedBox(height: AppSpacing.lg),
        Text('Mã đơn giao hàng', style: theme.textTheme.labelSmall),
        Text(trip.deliveryId.isEmpty ? trip.tripId : trip.deliveryId,
            style: theme.textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600)),
        const SizedBox(height: AppSpacing.lg),
        PickupCard(address: trip.pickupAddress.isEmpty ? 'Chưa cập nhật' : trip.pickupAddress),
        const RouteConnector(),
        DestinationCard(address: trip.dropoffAddress.isEmpty ? 'Chưa cập nhật' : trip.dropoffAddress),
        const SizedBox(height: AppSpacing.lg),
        AppCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Người nhận', style: theme.textTheme.titleSmall),
              const SizedBox(height: AppSpacing.sm),
              _Row(label: 'Tên', value: receiverName?.isNotEmpty == true ? receiverName : null),
              _Row(label: 'Số điện thoại', value: receiverPhone?.isNotEmpty == true ? receiverPhone : null),
            ],
          ),
        ),
        const SizedBox(height: AppSpacing.lg),
        AppCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Cước phí', style: theme.textTheme.titleSmall),
              const SizedBox(height: AppSpacing.sm),
              _Row(
                label: 'Giá ước tính',
                value: estimatedFareCents != null && estimatedFareCents! > 0 && currency != null
                    ? formatMoney(estimatedFareCents!, currency!)
                    : null,
                isTotal: true,
              ),
              const SizedBox(height: 4),
              Text(
                'Đây là giá ước tính khi đặt đơn — hệ thống tính phí giao hàng thực tế chưa khả dụng.',
                style: theme.textTheme.labelSmall?.copyWith(color: AppColors.textTertiary),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _Row extends StatelessWidget {
  const _Row({required this.label, required this.value, this.isTotal = false});

  final String label;
  final String? value;
  final bool isTotal;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final resolved = value ?? 'Chưa cập nhật';
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label,
              style: isTotal
                  ? theme.textTheme.titleSmall
                  : theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary)),
          Flexible(
            child: Text(
              resolved,
              textAlign: TextAlign.right,
              style: isTotal
                  ? theme.textTheme.titleMedium?.copyWith(color: AppColors.primary, fontWeight: FontWeight.w700)
                  : theme.textTheme.bodyMedium?.copyWith(
                      color: value == null ? AppColors.textTertiary : AppColors.textPrimary,
                      fontWeight: value == null ? FontWeight.w400 : FontWeight.w600,
                    ),
            ),
          ),
        ],
      ),
    );
  }
}
