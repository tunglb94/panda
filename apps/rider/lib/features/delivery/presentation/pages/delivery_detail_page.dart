import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/delivery/presentation/widgets/delivery_receipt_content.dart';
import 'package:rider/features/trip/data/trip_repository.dart';
import 'package:rider/shared/widgets/app_empty_state.dart';
import 'package:rider/shared/widgets/app_skeleton.dart';

/// Delivery order detail, opened from History's "Giao hàng" filter — a
/// separate page from `TripDetailPage` (Ride), rendering the Delivery-only
/// `DeliveryReceiptContent` rather than reusing Ride's receipt widget.
class DeliveryDetailPage extends StatefulWidget {
  const DeliveryDetailPage({super.key, required this.apiClient, required this.tripId});

  final ApiClient apiClient;
  final String tripId;

  @override
  State<DeliveryDetailPage> createState() => _DeliveryDetailPageState();
}

class _DeliveryDetailPageState extends State<DeliveryDetailPage> {
  late Future<TripDetail> _future;

  @override
  void initState() {
    super.initState();
    _future = TripRepository(widget.apiClient).getTrip(widget.tripId);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Chi tiết đơn giao hàng')),
      body: FutureBuilder<TripDetail>(
        future: _future,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return ListView(
              padding: const EdgeInsets.all(AppSpacing.lg),
              children: const [
                AppSkeletonBox(height: 140),
                SizedBox(height: AppSpacing.lg),
                AppSkeletonBox(height: 90),
              ],
            );
          }
          if (snap.hasError) {
            return AppEmptyState.error(
              subtitle: 'Không thể tải chi tiết đơn giao hàng. Vui lòng thử lại.',
              mascotAsset: 'mascot_no_connection.png',
              onAction: () => setState(() => _future = TripRepository(widget.apiClient).getTrip(widget.tripId)),
            );
          }
          return ListView(
            padding: const EdgeInsets.all(AppSpacing.lg),
            children: [DeliveryReceiptContent(trip: snap.data!)],
          );
        },
      ),
    );
  }
}
