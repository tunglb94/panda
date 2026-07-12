import 'package:flutter/material.dart';

import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/utils/currency_format.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_empty_state.dart';
import '../../../../shared/widgets/app_skeleton.dart';
import '../../../../shared/widgets/app_status_chip.dart';

class DriverTripHistoryPage extends StatefulWidget {
  const DriverTripHistoryPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<DriverTripHistoryPage> createState() => _DriverTripHistoryPageState();
}

enum _HistoryFilter { all, ride, delivery }

class _DriverTripHistoryPageState extends State<DriverTripHistoryPage> {
  late Future<List<_TripSummary>> _future;
  _HistoryFilter _filter = _HistoryFilter.all;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  List<_TripSummary> _applyFilter(List<_TripSummary> trips) => switch (_filter) {
        _HistoryFilter.all => trips,
        _HistoryFilter.ride => trips.where((t) => !t.isDelivery).toList(),
        _HistoryFilter.delivery => trips.where((t) => t.isDelivery).toList(),
      };

  Future<List<_TripSummary>> _load() async {
    final body = await widget.apiClient.get('/api/v1/driver/trips');
    final raw = (body['trips'] as List<dynamic>?) ?? [];
    return raw.map((e) => _TripSummary.fromJson(e as Map<String, dynamic>)).toList();
  }

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
            child: FutureBuilder<List<_TripSummary>>(
              future: _future,
              builder: (context, snap) {
                if (snap.connectionState == ConnectionState.waiting) {
                  return ListView.separated(
                    padding: const EdgeInsets.all(AppSpacing.lg),
                    itemCount: 5,
                    separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
                    itemBuilder: (context, i) => const AppSkeletonListTile(),
                  );
                }
                if (snap.hasError) {
                  return AppEmptyState.error(
                    subtitle: 'Không thể tải lịch sử chuyến đi. Vui lòng thử lại.',
                    mascotAsset: 'mascot_no_connection.png',
                    onAction: () => setState(() => _future = _load()),
                  );
                }
                final trips = _applyFilter(snap.data ?? []);
                if (trips.isEmpty) {
                  return AppEmptyState(
                    icon: Icons.receipt_long_outlined,
                    title: _filter == _HistoryFilter.delivery ? 'Chưa có đơn giao hàng nào' : 'Chưa có chuyến đi nào',
                    mascotAsset: 'mascot_waiting.png',
                  );
                }
                return RefreshIndicator(
                  onRefresh: () async => setState(() => _future = _load()),
                  child: ListView.separated(
                    padding: const EdgeInsets.all(AppSpacing.lg),
                    itemCount: trips.length,
                    separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
                    itemBuilder: (context, i) => _TripTile(trip: trips[i]),
                  ),
                );
              },
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
  const _TripTile({required this.trip});

  final _TripSummary trip;

  @override
  Widget build(BuildContext context) {
    final date = _formatDate(trip.createdAt);
    return AppCard(
      padding: const EdgeInsets.all(AppSpacing.md),
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
                    const SizedBox(width: 4),
                  ],
                  _StatusChip(status: trip.status),
                ],
              ),
              Text(date, style: Theme.of(context).textTheme.bodySmall),
            ],
          ),
          const SizedBox(height: AppSpacing.sm),
          _AddressRow(icon: Icons.location_on, color: AppColors.primary, label: trip.pickupAddress),
          const SizedBox(height: AppSpacing.xs),
          _AddressRow(icon: Icons.flag, color: AppColors.error, label: trip.dropoffAddress),
          if (trip.finalFare > 0) ...[
            const SizedBox(height: AppSpacing.sm),
            Align(
              alignment: Alignment.centerRight,
              child: Text(
                trip.fareText,
                style: Theme.of(context)
                    .textTheme
                    .titleSmall
                    ?.copyWith(color: AppColors.primary, fontSize: 15),
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
    final (label, color) = _info();
    return AppStatusChip(label: label, color: color);
  }

  (String, Color) _info() => switch (status) {
        'completed' || 'settled' => ('Hoàn thành', AppColors.primary),
        'cancelled' => ('Đã hủy', AppColors.error),
        'in_progress' => ('Đang thực hiện', AppColors.info),
        'payment_pending' || 'payment_success' => ('Đang chờ thanh toán', AppColors.warning),
        _ => ('Đang thực hiện', AppColors.textSecondary),
      };
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
        const SizedBox(width: 6),
        Expanded(
          child: Text(label, style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: AppColors.textPrimary,
              )),
        ),
      ],
    );
  }
}
