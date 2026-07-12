import '../../../core/network/api_client.dart';
import '../../../shared/utils/currency_format.dart';
import '../domain/models/driver_notification.dart';

/// Derives real [DriverNotification]s from the driver's actual trip history
/// (`GET /api/v1/driver/trips` — the same endpoint used elsewhere, no new
/// API): a settled trip becomes a "Thanh toán" notification with the real
/// amount; a cancelled trip becomes a "Chuyến xe"/"Giao hàng" notification;
/// a delivery trip whose `delivery_status` reached DELIVERED/COMPLETED
/// becomes a "Giao hàng" notification. Every other category (system/bonus/
/// promotion/update/warning/support) has no backend source anywhere in
/// this project, so this repository never returns entries for them — an
/// honest empty category rather than invented copy.
class NotificationRepository {
  const NotificationRepository(this._client);

  final ApiClient _client;

  Future<List<DriverNotification>> fetchAll() async {
    final body = await _client.get('/api/v1/driver/trips');
    final raw = (body['trips'] as List<dynamic>?) ?? [];
    final notifications = <DriverNotification>[];

    for (final e in raw) {
      final t = e as Map<String, dynamic>;
      final status = t['status'] as String? ?? '';
      DateTime dt;
      try {
        dt = DateTime.parse(t['created_at'] as String? ?? '').toLocal();
      } catch (_) {
        continue;
      }
      final tripId = t['trip_id'] as String? ?? '';
      final dropoff = t['dropoff_address'] as String? ?? '';
      final fareCents = (t['final_fare'] as num?)?.toInt() ?? 0;
      final currency = t['currency'] as String? ?? '';
      // Best-effort — see `enrichTripDetails` in the gateway's booking
      // handler. Empty trip_type means "ride" (the default).
      final isDelivery = t['trip_type'] == 'delivery';
      final deliveryStatus = t['delivery_status'] as String? ?? '';

      if (isDelivery && (deliveryStatus == 'DELIVERED' || deliveryStatus == 'COMPLETED')) {
        // Delivery never reaches "completed"/"settled" trip_status
        // (`CompleteDeliveryUseCase` deliberately leaves Trip.Status at
        // in_progress — no fare from Pricing to settle), so this is the
        // one real, honest signal for "delivered" — delivery_status.
        notifications.add(DriverNotification(
          id: 'delivery-$tripId',
          category: NotificationCategory.delivery,
          title: 'Đã giao hàng thành công',
          subtitle: 'Đơn giao đến $dropoff đã hoàn tất.',
          timestamp: dt,
          isRead: true,
        ));
      } else if (status == 'completed' || status == 'settled') {
        final amount = fareCents > 0 && currency.isNotEmpty ? formatMoney(fareCents, currency) : null;
        notifications.add(DriverNotification(
          id: 'payment-$tripId',
          category: NotificationCategory.payment,
          title: 'Đã nhận thanh toán',
          subtitle: amount != null
              ? 'Bạn đã nhận $amount cho chuyến đến $dropoff.'
              : 'Chuyến đến $dropoff đã hoàn tất.',
          timestamp: dt,
          isRead: true, // Historical trips — nothing to badge as "new".
        ));
      } else if (status == 'cancelled') {
        notifications.add(DriverNotification(
          id: 'trip-$tripId',
          category: isDelivery ? NotificationCategory.delivery : NotificationCategory.tripUpdate,
          title: isDelivery ? 'Đơn giao hàng đã hủy' : 'Chuyến đi đã hủy',
          subtitle: isDelivery ? 'Đơn giao đến $dropoff đã bị hủy.' : 'Chuyến đến $dropoff đã bị hủy.',
          timestamp: dt,
          isRead: true,
        ));
      }
    }

    notifications.sort((a, b) => b.timestamp.compareTo(a.timestamp));
    return notifications;
  }
}
