import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_status_chip.dart';
import '../../data/trip_offer_repository.dart';

/// Delivery offer card — a distinct widget from Ride's `_OfferCard`, per
/// the "Delivery must not reuse Ride's card" requirement. 📦 icon, item
/// type, receiver, distance, fee. `TripOffer` (from `GetDriverCurrentOffer`)
/// only carries `pickup_address`/`dropoff_address`/`trip_type` for a
/// delivery offer today — the Delivery entity's own fields (item
/// type/receiver name/declared value) are never exposed on any RPC a
/// reader can call (see the Delivery wire-contract audit), so those two
/// rows honestly show "Chưa cập nhật" rather than a guess. Distance/fee
/// are "—" for the same reason Ride's own offer card already shows "—"
/// for them — no estimate is computed pre-accept on either flow.
class DeliveryOfferCard extends StatelessWidget {
  const DeliveryOfferCard({
    super.key,
    required this.offer,
    required this.countdown,
    required this.onAccept,
    required this.onReject,
  });

  final TripOffer offer;
  final int countdown;
  final VoidCallback onAccept;
  final VoidCallback onReject;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isUrgent = countdown <= 10;
    return Padding(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: AppCard(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    const Icon(Icons.inventory_2, color: AppColors.info, size: AppIconSize.lg),
                    const SizedBox(width: AppSpacing.sm),
                    Text('Đơn giao hàng mới', style: theme.textTheme.titleLarge),
                  ],
                ),
                AppStatusChip(label: '${countdown}s', color: isUrgent ? AppColors.error : AppColors.info),
              ],
            ),
            const SizedBox(height: AppSpacing.lg),
            _InfoRow(icon: Icons.category_outlined, label: 'Loại hàng', value: 'Chưa cập nhật'),
            const SizedBox(height: AppSpacing.sm),
            _InfoRow(icon: Icons.person_outline, label: 'Người nhận', value: 'Chưa cập nhật'),
            const SizedBox(height: AppSpacing.lg),
            _AddressRow(icon: Icons.location_on, color: AppColors.primary, label: 'Điểm lấy hàng', address: offer.pickupAddress),
            const SizedBox(height: AppSpacing.md),
            _AddressRow(icon: Icons.flag, color: AppColors.error, label: 'Điểm giao hàng', address: offer.dropoffAddress),
            const SizedBox(height: AppSpacing.md),
            Row(
              children: const [
                _DeliveryInfoChip(icon: Icons.straighten, label: '—', sublabel: 'Khoảng cách'),
                SizedBox(width: AppSpacing.md),
                _DeliveryInfoChip(icon: Icons.attach_money, label: '—', sublabel: 'Phí giao hàng'),
              ],
            ),
            const SizedBox(height: AppSpacing.xxl),
            Row(
              children: [
                Expanded(child: AppButton.danger(label: 'Từ chối', onPressed: onReject)),
                const SizedBox(width: AppSpacing.md),
                Expanded(child: AppButton.primary(label: 'Chấp nhận', onPressed: onAccept)),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  const _InfoRow({required this.icon, required this.label, required this.value});

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Row(
      children: [
        Icon(icon, size: AppIconSize.sm, color: AppColors.textSecondary),
        const SizedBox(width: AppSpacing.sm),
        Text('$label: ', style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary)),
        Text(value, style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textTertiary)),
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

class _DeliveryInfoChip extends StatelessWidget {
  const _DeliveryInfoChip({required this.icon, required this.label, required this.sublabel});

  final IconData icon;
  final String label;
  final String sublabel;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
        decoration: BoxDecoration(color: AppColors.surfaceAlt, borderRadius: AppRadius.smAll),
        child: Row(
          children: [
            Icon(icon, size: AppIconSize.sm, color: AppColors.textSecondary),
            const SizedBox(width: AppSpacing.sm),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(label, style: theme.textTheme.titleSmall),
                Text(sublabel, style: theme.textTheme.labelSmall?.copyWith(color: AppColors.textTertiary)),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
