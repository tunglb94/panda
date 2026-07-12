import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/mascot_image.dart';
import '../../data/active_trip_repository.dart';
import '../../domain/models/delivery_status.dart';

/// Delivery lifecycle card — a distinct widget from Ride's
/// `_TripExecutionCard`/`_AwaitingPaymentCard`/`_TripCompletedCard`. Drives
/// its sub-state off `ActiveTrip.deliveryStatus` (not `trip_status` alone —
/// see `DeliveryStatus`'s doc comment) through the full Accept→Arrive
/// Pickup→Pickup Parcel→Start Delivery→Complete Delivery lifecycle in one
/// self-contained widget, since Delivery has no payment/rating step to
/// hand off to (`_PageState` stays `activeTrip` throughout).
class DeliveryExecutionCard extends StatelessWidget {
  const DeliveryExecutionCard({
    super.key,
    required this.trip,
    required this.hasArrived,
    required this.onArrived,
    required this.onPickupParcel,
    required this.onStartDelivery,
    required this.onCompleteDelivery,
    required this.onDone,
  });

  final ActiveTrip trip;
  final bool hasArrived;
  final VoidCallback onArrived;
  final VoidCallback onPickupParcel;
  final VoidCallback onStartDelivery;
  final VoidCallback onCompleteDelivery;
  final VoidCallback onDone;

  DeliveryStatus get _status => DeliveryStatus.fromWire(trip.deliveryStatus);

  @override
  Widget build(BuildContext context) {
    if (_status == DeliveryStatus.delivered || _status == DeliveryStatus.completed) {
      return _CompletedView(onDone: onDone);
    }

    final theme = Theme.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Column(
        children: [
          _DeliveryTimeline(status: _status, hasArrived: hasArrived),
          const SizedBox(height: AppSpacing.lg),
          AppCard(
            padding: const EdgeInsets.all(AppSpacing.xl),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Icon(Icons.inventory_2, color: AppColors.info, size: AppIconSize.lg),
                    const SizedBox(width: AppSpacing.sm),
                    Text('Đơn giao hàng', style: theme.textTheme.titleLarge),
                  ],
                ),
                const SizedBox(height: AppSpacing.lg),
                _AddressRow(
                  icon: Icons.location_on,
                  color: AppColors.primary,
                  label: 'Điểm lấy hàng',
                  address: trip.pickupAddress,
                ),
                const SizedBox(height: AppSpacing.md),
                _AddressRow(
                  icon: Icons.flag,
                  color: AppColors.error,
                  label: 'Điểm giao hàng',
                  address: trip.dropoffAddress,
                ),
                const SizedBox(height: AppSpacing.md),
                Row(
                  children: [
                    Icon(Icons.person_outline, size: AppIconSize.sm, color: AppColors.textSecondary),
                    const SizedBox(width: AppSpacing.sm),
                    Text('Người nhận: ', style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary)),
                    Text('Chưa cập nhật', style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textTertiary)),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: AppSpacing.xl),
          _ActionButton(
            status: _status,
            hasArrived: hasArrived,
            onArrived: onArrived,
            onPickupParcel: onPickupParcel,
            onStartDelivery: onStartDelivery,
            onCompleteDelivery: onCompleteDelivery,
          ),
        ],
      ),
    );
  }
}

class _ActionButton extends StatelessWidget {
  const _ActionButton({
    required this.status,
    required this.hasArrived,
    required this.onArrived,
    required this.onPickupParcel,
    required this.onStartDelivery,
    required this.onCompleteDelivery,
  });

  final DeliveryStatus status;
  final bool hasArrived;
  final VoidCallback onArrived;
  final VoidCallback onPickupParcel;
  final VoidCallback onStartDelivery;
  final VoidCallback onCompleteDelivery;

  @override
  Widget build(BuildContext context) {
    if (status == DeliveryStatus.inDelivery) {
      return AppButton.danger(label: 'Đã giao hàng', onPressed: onCompleteDelivery);
    }
    if (status == DeliveryStatus.parcelPickedUp) {
      return AppButton.primary(label: 'Bắt đầu giao hàng', onPressed: onStartDelivery);
    }
    // created/accepted/unknown — mirrors Ride's arrive flow (Trip's
    // arrive/MarkDriverArrived is reused unchanged for Delivery).
    if (hasArrived) {
      return AppButton.primary(label: 'Xác nhận đã lấy hàng', onPressed: onPickupParcel);
    }
    return AppButton.outline(label: 'Tôi đã đến điểm lấy hàng', onPressed: onArrived);
  }
}

class _CompletedView extends StatelessWidget {
  const _CompletedView({required this.onDone});

  final VoidCallback onDone;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const MascotImage(
              asset: 'mascot_celebration.png',
              size: MascotSize.large,
              animation: MascotAnimation.bounce,
              semanticLabel: 'Giao hàng thành công',
            ),
            const SizedBox(height: AppSpacing.md),
            Text('Giao hàng thành công!',
                style: theme.textTheme.headlineSmall?.copyWith(color: AppColors.primary, fontWeight: FontWeight.bold)),
            const SizedBox(height: AppSpacing.xxl),
            AppButton.primary(label: 'Quay lại hàng đợi', onPressed: onDone),
          ],
        ),
      ),
    );
  }
}

class _DeliveryTimeline extends StatelessWidget {
  const _DeliveryTimeline({required this.status, required this.hasArrived});

  final DeliveryStatus status;
  final bool hasArrived;

  int get _stepIndex {
    if (status == DeliveryStatus.inDelivery) return 2;
    if (status == DeliveryStatus.parcelPickedUp) return 1;
    if (hasArrived) return 1;
    return 0;
  }

  static const _labels = ['Đến điểm lấy hàng', 'Đã lấy hàng', 'Đang giao', 'Hoàn thành'];

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final current = _stepIndex;
    return Row(
      children: [
        for (var i = 0; i < _labels.length; i++) ...[
          Expanded(
            child: Column(
              children: [
                Container(
                  width: 12,
                  height: 12,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: i <= current ? AppColors.primary : AppColors.border,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  _labels[i],
                  textAlign: TextAlign.center,
                  style: theme.textTheme.labelSmall?.copyWith(
                    color: i <= current ? AppColors.primary : AppColors.textTertiary,
                    fontWeight: i == current ? FontWeight.w700 : FontWeight.w400,
                  ),
                ),
              ],
            ),
          ),
          if (i != _labels.length - 1)
            Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: Container(width: 16, height: 2, color: i < current ? AppColors.primary : AppColors.border),
            ),
        ],
      ],
    );
  }
}

class _AddressRow extends StatelessWidget {
  const _AddressRow({required this.icon, required this.color, required this.label, required this.address});

  final IconData icon;
  final Color color;
  final String label;
  final String address;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, color: color, size: AppIconSize.md),
        const SizedBox(width: AppSpacing.sm),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label, style: theme.textTheme.labelSmall),
              Text(address, style: theme.textTheme.bodyMedium),
            ],
          ),
        ),
      ],
    );
  }
}
