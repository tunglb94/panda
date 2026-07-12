import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/trip/data/trip_repository.dart';
import 'package:rider/shared/widgets/app_bottom_sheet.dart';
import 'package:rider/shared/widgets/app_empty_state.dart';
import 'package:rider/shared/widgets/app_skeleton.dart';

import 'delivery_receipt_content.dart';

/// "Xem hóa đơn" on the Delivery success screen — re-fetches the trip fresh
/// so the receipt is correct even if opened a while after delivery, mirrors
/// `TripReceiptSheet`'s shape but is a wholly separate widget per the
/// "Delivery Receipt must not reuse Ride's" requirement.
abstract final class DeliveryReceiptSheet {
  static Future<void> show(
    BuildContext context, {
    required String tripId,
    required ApiClient apiClient,
    String? receiverName,
    String? receiverPhone,
    int? estimatedFareCents,
    String? currency,
  }) {
    return AppBottomSheet.show<void>(
      context,
      title: 'Hóa đơn giao hàng',
      isScrollControlled: true,
      builder: (_) => _DeliveryReceiptBody(
        tripId: tripId,
        apiClient: apiClient,
        receiverName: receiverName,
        receiverPhone: receiverPhone,
        estimatedFareCents: estimatedFareCents,
        currency: currency,
      ),
    );
  }
}

class _DeliveryReceiptBody extends StatefulWidget {
  const _DeliveryReceiptBody({
    required this.tripId,
    required this.apiClient,
    this.receiverName,
    this.receiverPhone,
    this.estimatedFareCents,
    this.currency,
  });

  final String tripId;
  final ApiClient apiClient;
  final String? receiverName;
  final String? receiverPhone;
  final int? estimatedFareCents;
  final String? currency;

  @override
  State<_DeliveryReceiptBody> createState() => _DeliveryReceiptBodyState();
}

class _DeliveryReceiptBodyState extends State<_DeliveryReceiptBody> {
  late Future<TripDetail> _future;

  @override
  void initState() {
    super.initState();
    _future = TripRepository(widget.apiClient).getTrip(widget.tripId);
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<TripDetail>(
      future: _future,
      builder: (context, snap) {
        if (snap.connectionState == ConnectionState.waiting) {
          return const Padding(
            padding: EdgeInsets.symmetric(vertical: AppSpacing.xl),
            child: Column(
              children: [
                AppSkeletonBox(height: 90),
                SizedBox(height: AppSpacing.md),
                AppSkeletonBox(height: 120),
              ],
            ),
          );
        }
        if (snap.hasError) {
          return AppEmptyState.error(
            subtitle: 'Không thể tải hóa đơn. Vui lòng thử lại.',
            mascotAsset: 'mascot_no_connection.png',
            onAction: () => setState(() => _future = TripRepository(widget.apiClient).getTrip(widget.tripId)),
          );
        }
        return DeliveryReceiptContent(
          trip: snap.data!,
          receiverName: widget.receiverName,
          receiverPhone: widget.receiverPhone,
          estimatedFareCents: widget.estimatedFareCents,
          currency: widget.currency,
        );
      },
    );
  }
}
