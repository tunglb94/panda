import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/trip/data/trip_repository.dart';
import 'package:rider/features/trip/domain/models/driver_profile.dart';
import 'package:rider/shared/widgets/app_bottom_sheet.dart';
import 'package:rider/shared/widgets/app_empty_state.dart';
import 'package:rider/shared/widgets/app_skeleton.dart';

import 'trip_receipt_content.dart';

/// "Xem hóa đơn" — the Receipt CTA on the Payment Success screen (Section 8)
/// opens this. It re-fetches the trip fresh (rather than trusting only the
/// in-memory poll state) so the receipt is correct even if the rider opens
/// it a while after payment.
abstract final class TripReceiptSheet {
  static Future<void> show(
    BuildContext context, {
    required String tripId,
    required ApiClient apiClient,
    String? paymentMethodLabel,
  }) {
    return AppBottomSheet.show<void>(
      context,
      title: 'Hóa đơn chuyến đi',
      isScrollControlled: true,
      builder: (_) => _TripReceiptBody(
        tripId: tripId,
        apiClient: apiClient,
        paymentMethodLabel: paymentMethodLabel,
      ),
    );
  }
}

class _ReceiptData {
  const _ReceiptData({required this.trip, this.vehicle});
  final TripDetail trip;
  final DriverProfile? vehicle;
}

class _TripReceiptBody extends StatefulWidget {
  const _TripReceiptBody({
    required this.tripId,
    required this.apiClient,
    this.paymentMethodLabel,
  });

  final String tripId;
  final ApiClient apiClient;
  final String? paymentMethodLabel;

  @override
  State<_TripReceiptBody> createState() => _TripReceiptBodyState();
}

class _TripReceiptBodyState extends State<_TripReceiptBody> {
  late Future<_ReceiptData> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<_ReceiptData> _load() async {
    final repo = TripRepository(widget.apiClient);
    final trip = await repo.getTrip(widget.tripId);
    DriverProfile? vehicle;
    if (trip.driverId.isNotEmpty) {
      try {
        vehicle = await repo.fetchDriverProfile(trip.driverId);
      } on ApiException {
        vehicle = null;
      }
    }
    return _ReceiptData(trip: trip, vehicle: vehicle);
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<_ReceiptData>(
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
            onAction: () => setState(() => _future = _load()),
          );
        }
        return TripReceiptContent(
          trip: snap.data!.trip,
          vehicle: snap.data!.vehicle,
          paymentMethodLabel: widget.paymentMethodLabel,
        );
      },
    );
  }
}
