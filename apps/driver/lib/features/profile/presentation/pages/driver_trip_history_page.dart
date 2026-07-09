import 'package:flutter/material.dart';

import '../../../../core/network/api_client.dart';

class DriverTripHistoryPage extends StatefulWidget {
  const DriverTripHistoryPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<DriverTripHistoryPage> createState() => _DriverTripHistoryPageState();
}

class _DriverTripHistoryPageState extends State<DriverTripHistoryPage> {
  late Future<List<_TripSummary>> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<_TripSummary>> _load() async {
    final body = await widget.apiClient.get('/api/v1/driver/trips');
    final raw = (body['trips'] as List<dynamic>?) ?? [];
    return raw.map((e) => _TripSummary.fromJson(e as Map<String, dynamic>)).toList();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Trip History')),
      body: FutureBuilder<List<_TripSummary>>(
        future: _future,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snap.hasError) {
            return _ErrorView(
              message: snap.error is ApiException
                  ? (snap.error as ApiException).message
                  : 'Failed to load history.',
              onRetry: () => setState(() => _future = _load()),
            );
          }
          final trips = snap.data ?? [];
          if (trips.isEmpty) {
            return const Center(
              child: Text(
                'No trips yet.',
                style: TextStyle(color: Colors.grey),
              ),
            );
          }
          return RefreshIndicator(
            onRefresh: () async => setState(() => _future = _load()),
            child: ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: trips.length,
              separatorBuilder: (_, __) => const SizedBox(height: 8),
              itemBuilder: (context, i) => _TripTile(trip: trips[i]),
            ),
          );
        },
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
  });

  final String tripId;
  final String status;
  final String pickupAddress;
  final String dropoffAddress;
  final int finalFare;
  final String currency;
  final DateTime createdAt;

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
    );
  }

  String get fareText {
    if (finalFare <= 0 || currency.isEmpty) return '—';
    final sym = currency.toUpperCase() == 'USD' ? r'$' : currency;
    return '$sym${(finalFare / 100).toStringAsFixed(2)}';
  }
}

class _TripTile extends StatelessWidget {
  const _TripTile({required this.trip});

  final _TripSummary trip;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final date = _formatDate(trip.createdAt);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                _StatusChip(status: trip.status),
                Text(date, style: TextStyle(fontSize: 12, color: cs.onSurfaceVariant)),
              ],
            ),
            const SizedBox(height: 10),
            _AddressRow(
              icon: Icons.location_on,
              color: cs.primary,
              label: trip.pickupAddress,
            ),
            const SizedBox(height: 6),
            _AddressRow(
              icon: Icons.flag,
              color: cs.error,
              label: trip.dropoffAddress,
            ),
            if (trip.finalFare > 0) ...[
              const SizedBox(height: 10),
              Align(
                alignment: Alignment.centerRight,
                child: Text(
                  trip.fareText,
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 15,
                    color: cs.primary,
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  static String _formatDate(DateTime dt) {
    final now = DateTime.now();
    if (dt.year == now.year && dt.month == now.month && dt.day == now.day) {
      return 'Today ${_hhmm(dt)}';
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
    final cs = Theme.of(context).colorScheme;
    final (label, color) = _info(cs);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        label,
        style: TextStyle(fontSize: 12, color: color, fontWeight: FontWeight.w600),
      ),
    );
  }

  (String, Color) _info(ColorScheme cs) => switch (status) {
        'completed' || 'settled' => ('Completed', cs.primary),
        'cancelled' => ('Cancelled', cs.error),
        'in_progress' => ('In Progress', cs.tertiary),
        'payment_pending' || 'payment_success' => ('Payment Pending', cs.secondary),
        _ => ('In Progress', cs.onSurfaceVariant),
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
        Icon(icon, size: 16, color: color),
        const SizedBox(width: 6),
        Expanded(
          child: Text(label, style: const TextStyle(fontSize: 13)),
        ),
      ],
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.error_outline, size: 48, color: Theme.of(context).colorScheme.error),
            const SizedBox(height: 12),
            Text(message, textAlign: TextAlign.center),
            const SizedBox(height: 16),
            FilledButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }
}
