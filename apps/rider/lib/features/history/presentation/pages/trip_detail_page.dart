import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/trip/data/trip_repository.dart';
import 'package:rider/features/trip/domain/models/driver_profile.dart';
import 'package:rider/features/trip/presentation/widgets/trip_receipt_content.dart';
import 'package:rider/shared/widgets/app_empty_state.dart';
import 'package:rider/shared/widgets/app_skeleton.dart';

class _DetailData {
  const _DetailData({required this.trip, this.vehicle});
  final TripDetail trip;
  final DriverProfile? vehicle;
}

/// Trip detail / receipt — a real, honest rebuild replacing the old
/// mock-catalog-only `TripDetailPage`/`ReceiptPage` (dead code, never
/// reachable from any route, built against fields the backend doesn't
/// return: itemized tax, named driver photo, per-status timeline
/// timestamps). This version only shows what `GET /api/v1/rides/{tripId}`
/// and `GET /api/v1/drivers/{driverId}/profile` actually return — the body
/// is [TripReceiptContent], the same widget the post-payment "Xem hóa đơn"
/// CTA opens (Section 9 of the Payment/Fare production pass), so a trip's
/// receipt looks identical whether opened right after paying or later from
/// history. No itemized breakdown, no timeline, no fabricated names —
/// fields the backend doesn't provide show "Chưa cập nhật" rather than
/// being invented.
class TripDetailPage extends StatefulWidget {
  const TripDetailPage({
    super.key,
    required this.apiClient,
    required this.tripId,
    required this.pickupAddress,
    required this.dropoffAddress,
    required this.createdAt,
  });

  final ApiClient apiClient;
  final String tripId;
  final String pickupAddress;
  final String dropoffAddress;
  final DateTime createdAt;

  @override
  State<TripDetailPage> createState() => _TripDetailPageState();
}

class _TripDetailPageState extends State<TripDetailPage> {
  late Future<_DetailData> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<_DetailData> _load() async {
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
    return _DetailData(trip: trip, vehicle: vehicle);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Chi tiết chuyến đi')),
      body: FutureBuilder<_DetailData>(
        future: _future,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return ListView(
              padding: const EdgeInsets.all(AppSpacing.lg),
              children: const [
                AppSkeletonBox(height: 140),
                SizedBox(height: AppSpacing.lg),
                AppSkeletonBox(height: 90),
                SizedBox(height: AppSpacing.lg),
                AppSkeletonBox(height: 90),
              ],
            );
          }
          if (snap.hasError) {
            return AppEmptyState.error(
              subtitle: 'Không thể tải chi tiết chuyến đi. Vui lòng thử lại.',
              mascotAsset: 'mascot_no_connection.png',
              onAction: () => setState(() => _future = _load()),
            );
          }

          // The list tile that navigated here already knows the addresses
          // and created_at; prefer the freshly-fetched trip's own
          // pickup/dropoff (now parsed by TripRepository.getTrip) and fall
          // back to the navigation args only if the backend omitted them.
          final trip = snap.data!.trip;
          final tripWithFallback = trip.pickupAddress.isNotEmpty || trip.dropoffAddress.isNotEmpty
              ? trip
              : TripDetail(
                  tripId: trip.tripId,
                  status: trip.status,
                  driverId: trip.driverId,
                  finalFareCents: trip.finalFareCents,
                  currency: trip.currency,
                  pickupAddress: widget.pickupAddress,
                  dropoffAddress: widget.dropoffAddress,
                  dispatchStatus: trip.dispatchStatus,
                );

          return ListView(
            padding: const EdgeInsets.all(AppSpacing.lg),
            children: [
              TripReceiptContent(
                trip: tripWithFallback,
                vehicle: snap.data!.vehicle,
                createdAt: widget.createdAt,
              ),
            ],
          );
        },
      ),
    );
  }
}
