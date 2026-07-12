import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/delivery/presentation/pages/delivery_detail_page.dart';
import 'package:rider/shared/widgets/app_card.dart';
import 'package:rider/shared/widgets/app_empty_state.dart';
import 'package:rider/shared/widgets/app_skeleton.dart';
import 'package:rider/shared/utils/currency_format.dart';
import 'package:rider/shared/widgets/app_status_chip.dart';
import 'package:rider/shared/widgets/async_state_view.dart';

import 'trip_detail_page.dart';

class TripHistoryPage extends StatefulWidget {
  const TripHistoryPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<TripHistoryPage> createState() => _TripHistoryPageState();
}

enum _HistoryFilter { all, ride, delivery }

class _TripHistoryPageState extends State<TripHistoryPage> {
  late Future<List<_TripSummary>> _future;
  _HistoryFilter _filter = _HistoryFilter.all;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<_TripSummary>> _load() async {
    final body = await widget.apiClient.get('/api/v1/rider/trips');
    final raw = (body['trips'] as List<dynamic>?) ?? [];
    return raw.map((e) => _TripSummary.fromJson(e as Map<String, dynamic>)).toList();
  }

  List<_TripSummary> _applyFilter(List<_TripSummary> trips) => switch (_filter) {
        _HistoryFilter.all => trips,
        _HistoryFilter.ride => trips.where((t) => !t.isDelivery).toList(),
        _HistoryFilter.delivery => trips.where((t) => t.isDelivery).toList(),
      };

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Lịch sử chuyến đi')),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg, vertical: AppSpacing.sm),
            child: Row(
              children: [
                ChoiceChip(
                  label: const Text('Tất cả'),
                  selected: _filter == _HistoryFilter.all,
                  onSelected: (_) => setState(() => _filter = _HistoryFilter.all),
                ),
                const SizedBox(width: AppSpacing.sm),
                ChoiceChip(
                  label: const Text('Chuyến xe'),
                  selected: _filter == _HistoryFilter.ride,
                  onSelected: (_) => setState(() => _filter = _HistoryFilter.ride),
                ),
                const SizedBox(width: AppSpacing.sm),
                ChoiceChip(
                  label: const Text('Giao hàng'),
                  selected: _filter == _HistoryFilter.delivery,
                  onSelected: (_) => setState(() => _filter = _HistoryFilter.delivery),
                ),
              ],
            ),
          ),
          Expanded(
            child: RefreshIndicator(
              onRefresh: () async => setState(() => _future = _load()),
              child: AsyncStateView<List<_TripSummary>>(
                future: _future,
                isEmpty: (trips) => _applyFilter(trips).isEmpty,
                loadingBuilder: (context) => ListView(
                  padding: const EdgeInsets.all(AppSpacing.lg),
                  children: const [
                    AppSkeletonListTile(),
                    SizedBox(height: AppSpacing.sm),
                    AppSkeletonListTile(),
                    SizedBox(height: AppSpacing.sm),
                    AppSkeletonListTile(),
                  ],
                ),
                emptyBuilder: (context) => AppEmptyState(
                  icon: Icons.receipt_long_outlined,
                  title: _filter == _HistoryFilter.delivery ? 'Chưa có đơn giao hàng nào' : 'Chưa có chuyến đi nào',
                  subtitle: 'Các chuyến đã hoàn tất sẽ xuất hiện ở đây.',
                  mascotAsset: 'mascot_waiting.png',
                ),
                errorBuilder: (context, error) => AppEmptyState.error(
                  subtitle: error is ApiException && error.statusCode == 0
                      ? error.message
                      : 'Không thể tải lịch sử chuyến đi.',
                  onAction: () => setState(() => _future = _load()),
                  mascotAsset: 'mascot_no_connection.png',
                ),
                successBuilder: (context, allTrips) {
                  final trips = _applyFilter(allTrips);
                  return ListView.separated(
                    padding: const EdgeInsets.all(AppSpacing.lg),
                    itemCount: trips.length,
                    separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
                    itemBuilder: (context, i) => _TripTile(
                      trip: trips[i],
                      onTap: () => Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (_) => trips[i].isDelivery
                              ? DeliveryDetailPage(apiClient: widget.apiClient, tripId: trips[i].tripId)
                              : TripDetailPage(
                                  apiClient: widget.apiClient,
                                  tripId: trips[i].tripId,
                                  pickupAddress: trips[i].pickupAddress,
                                  dropoffAddress: trips[i].dropoffAddress,
                                  createdAt: trips[i].createdAt,
                                ),
                        ),
                      ),
                    ),
                  );
                },
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _TripSummary {
  const _TripSummary({
    required this.tripId,
    required this.status,
    required this.pickupAddress,
    required this.dropoffAddress,
    required this.finalFare,
    required this.currency,
    required this.createdAt,
    this.tripType = '',
  });

  final String tripId;
  final String status;
  final String pickupAddress;
  final String dropoffAddress;
  final int finalFare;
  final String currency;
  final DateTime createdAt;

  /// Best-effort — see `enrichTripTypes` in the gateway's booking handler.
  /// Empty means "ride" (the default) or the per-trip lookup failed.
  final String tripType;

  bool get isDelivery => tripType == 'delivery';

  factory _TripSummary.fromJson(Map<String, dynamic> j) {
    DateTime dt;
    try {
      dt = DateTime.parse(j['created_at'] as String? ?? '');
    } catch (_) {
      dt = DateTime.fromMillisecondsSinceEpoch(0);
    }
    return _TripSummary(
      tripId: j['trip_id'] as String? ?? '',
      status: j['status'] as String? ?? '',
      pickupAddress: j['pickup_address'] as String? ?? '',
      dropoffAddress: j['dropoff_address'] as String? ?? '',
      finalFare: (j['final_fare'] as num?)?.toInt() ?? 0,
      currency: j['currency'] as String? ?? '',
      createdAt: dt.toLocal(),
      tripType: j['trip_type'] as String? ?? '',
    );
  }

  String get fareText {
    if (finalFare <= 0 || currency.isEmpty) return '—';
    return formatMoney(finalFare, currency);
  }
}

class _TripTile extends StatelessWidget {
  const _TripTile({required this.trip, required this.onTap});

  final _TripSummary trip;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final date = _formatDate(trip.createdAt);
    return AppCard(
      padding: const EdgeInsets.all(AppSpacing.md),
      onTap: onTap,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (trip.isDelivery) ...[
                    const Icon(Icons.local_shipping_outlined, size: AppIconSize.sm, color: AppColors.info),
                    const SizedBox(width: AppSpacing.xs),
                  ],
                  _StatusChip(status: trip.status),
                ],
              ),
              Text(date, style: Theme.of(context).textTheme.bodySmall),
            ],
          ),
          const SizedBox(height: AppSpacing.md),
          _AddressRow(
            icon: Icons.my_location,
            color: AppColors.primary,
            label: trip.pickupAddress,
          ),
          const SizedBox(height: AppSpacing.xs),
          _AddressRow(
            icon: Icons.flag,
            color: AppColors.error,
            label: trip.dropoffAddress,
          ),
          if (trip.finalFare > 0) ...[
            const SizedBox(height: AppSpacing.sm),
            Align(
              alignment: Alignment.centerRight,
              child: Text(
                trip.fareText,
                style: Theme.of(context).textTheme.titleSmall?.copyWith(color: AppColors.primary),
              ),
            ),
          ],
        ],
      ),
    );
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

class _StatusChip extends StatelessWidget {
  const _StatusChip({required this.status});

  final String status;

  @override
  Widget build(BuildContext context) {
    final (label, color) = switch (status) {
      'completed' || 'settled' => ('Hoàn tất', AppColors.primary),
      'cancelled' => ('Đã hủy', AppColors.error),
      'in_progress' => ('Đang di chuyển', AppColors.info),
      'payment_pending' || 'payment_success' => ('Chờ thanh toán', AppColors.warning),
      _ => ('Đang xử lý', AppColors.textTertiary),
    };
    return AppStatusChip(label: label, color: color);
  }
}

class _AddressRow extends StatelessWidget {
  const _AddressRow({required this.icon, required this.color, required this.label});

  final IconData icon;
  final Color color;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, size: AppIconSize.sm, color: color),
        const SizedBox(width: AppSpacing.sm),
        Expanded(
          child: Text(label, style: Theme.of(context).textTheme.bodySmall),
        ),
      ],
    );
  }
}
