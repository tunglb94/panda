import 'package:flutter/material.dart';

import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/booking/presentation/widgets/trip_point_cards.dart';
import 'package:rider/features/trip/data/trip_repository.dart';
import 'package:rider/features/trip/domain/models/driver_profile.dart';
import 'package:rider/shared/utils/currency_format.dart';
import 'package:rider/shared/widgets/app_card.dart';
import 'package:rider/shared/widgets/app_status_chip.dart';

/// Full receipt content — production-polish pass. Shared by
/// [TripReceiptSheet] (opened right after payment, from the Payment
/// Success screen) and `TripDetailPage` (opened later from trip history),
/// so the exact same real-vs-"Chưa cập nhật" rules apply no matter which
/// entry point the rider used.
///
/// Every row here is either a real field already returned by
/// `GET /api/v1/rides/{tripId}` / `GET /api/v1/drivers/{driverId}/profile`,
/// or an explicit "Chưa cập nhật" placeholder for a field the backend does
/// not expose today (distance, duration, vehicle type, driver name, VAT,
/// platform fee, voucher, promotion — see the Payment/Fare production-pass
/// audit). Nothing is computed or estimated here; that would misstate a
/// document the rider may treat as a real receipt.
class TripReceiptContent extends StatelessWidget {
  const TripReceiptContent({
    super.key,
    required this.trip,
    this.vehicle,
    this.paymentMethodLabel,
    this.createdAt,
  });

  final TripDetail trip;
  final DriverProfile? vehicle;

  /// 'tiền mặt'/'ví điện tử' — only known for the trip just paid in this
  /// app session (the backend never echoes payment method back on any
  /// response). Null for a trip opened later from history.
  final String? paymentMethodLabel;

  /// Only available when the caller already has it (e.g. passed down from
  /// a trip-history list tile). `GetTrip`'s response has no timestamp
  /// field, so a receipt opened right after payment cannot show one.
  final DateTime? createdAt;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Row(
          children: [
            _statusChip(trip.status),
            const Spacer(),
            Text(createdAt != null ? _formatDate(createdAt!) : 'Chưa cập nhật', style: theme.textTheme.bodySmall),
          ],
        ),
        const SizedBox(height: AppSpacing.lg),
        Text('Mã chuyến đi', style: theme.textTheme.labelSmall),
        Text(trip.tripId, style: theme.textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600)),
        const SizedBox(height: AppSpacing.lg),
        PickupCard(address: trip.pickupAddress.isEmpty ? 'Chưa cập nhật' : trip.pickupAddress),
        const RouteConnector(),
        DestinationCard(address: trip.dropoffAddress.isEmpty ? 'Chưa cập nhật' : trip.dropoffAddress),
        const SizedBox(height: AppSpacing.lg),
        AppCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Chi tiết cước phí', style: theme.textTheme.titleSmall),
              const SizedBox(height: AppSpacing.sm),
              const _ReceiptRow(label: 'Khoảng cách', value: null),
              const _ReceiptRow(label: 'Thời gian', value: null),
              const _ReceiptRow(label: 'Voucher', value: null),
              const _ReceiptRow(label: 'Khuyến mãi', value: null),
              const _ReceiptRow(label: 'Phí nền tảng', value: null),
              const _ReceiptRow(label: 'VAT', value: null),
              const Divider(height: AppSpacing.lg, color: AppColors.divider),
              _ReceiptRow(
                label: 'Tổng thanh toán',
                value: trip.finalFareCents > 0 && trip.currency.isNotEmpty
                    ? formatMoney(trip.finalFareCents, trip.currency)
                    : null,
                isTotal: true,
              ),
            ],
          ),
        ),
        const SizedBox(height: AppSpacing.lg),
        AppCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Phương tiện & thanh toán', style: theme.textTheme.titleSmall),
              const SizedBox(height: AppSpacing.sm),
              const _ReceiptRow(label: 'Loại xe', value: null),
              _ReceiptRow(
                label: 'Xe',
                value: vehicle != null && vehicle!.vehicleDisplay.isNotEmpty ? vehicle!.vehicleDisplay : null,
              ),
              _ReceiptRow(
                label: 'Biển số',
                value: vehicle != null && vehicle!.plateNumber.isNotEmpty && vehicle!.plateNumber != '—'
                    ? vehicle!.plateNumber
                    : null,
              ),
              const _ReceiptRow(label: 'Tên tài xế', value: null),
              _ReceiptRow(label: 'Phương thức thanh toán', value: paymentMethodLabel),
            ],
          ),
        ),
      ],
    );
  }

  Widget _statusChip(String status) {
    final (label, color) = switch (status) {
      'completed' || 'settled' || 'payment_success' => ('Hoàn tất', AppColors.primary),
      'cancelled' => ('Đã hủy', AppColors.error),
      'in_progress' => ('Đang di chuyển', AppColors.info),
      'payment_pending' => ('Chờ thanh toán', AppColors.warning),
      _ => ('Đang xử lý', AppColors.textTertiary),
    };
    return AppStatusChip(label: label, color: color);
  }

  static String _formatDate(DateTime dt) {
    final now = DateTime.now();
    if (dt.year == now.year && dt.month == now.month && dt.day == now.day) {
      return 'Hôm nay ${_hhmm(dt)}';
    }
    return '${dt.day}/${dt.month}/${dt.year} ${_hhmm(dt)}';
  }

  static String _hhmm(DateTime dt) =>
      '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
}

class _ReceiptRow extends StatelessWidget {
  const _ReceiptRow({required this.label, required this.value, this.isTotal = false});

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
          Text(
            label,
            style: isTotal
                ? theme.textTheme.titleSmall
                : theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
          ),
          Text(
            resolved,
            style: isTotal
                ? theme.textTheme.titleMedium?.copyWith(color: AppColors.primary, fontWeight: FontWeight.w700)
                : theme.textTheme.bodyMedium?.copyWith(
                    color: value == null ? AppColors.textTertiary : AppColors.textPrimary,
                    fontWeight: value == null ? FontWeight.w400 : FontWeight.w600,
                  ),
          ),
        ],
      ),
    );
  }
}
