/// Mirrors the rider app's `DeliveryStatus` — both parse the exact same
/// wire values from `backend/services/trip/domain/entity/delivery.go`'s
/// `DeliveryStatus` enum (CREATED/ACCEPTED/PARCEL_PICKED_UP/IN_DELIVERY/
/// DELIVERED/COMPLETED/CANCELLED). Kept as a separate copy rather than a
/// shared package since the two apps have no shared Dart package today.
enum DeliveryStatus {
  created,
  accepted,
  parcelPickedUp,
  inDelivery,
  delivered,
  completed,
  cancelled,
  unknown;

  static DeliveryStatus fromWire(String raw) => switch (raw) {
        'CREATED' => DeliveryStatus.created,
        'ACCEPTED' => DeliveryStatus.accepted,
        'PARCEL_PICKED_UP' => DeliveryStatus.parcelPickedUp,
        'IN_DELIVERY' => DeliveryStatus.inDelivery,
        'DELIVERED' => DeliveryStatus.delivered,
        'COMPLETED' => DeliveryStatus.completed,
        'CANCELLED' => DeliveryStatus.cancelled,
        _ => DeliveryStatus.unknown,
      };

  bool get isTerminal => this == DeliveryStatus.completed || this == DeliveryStatus.cancelled;

  String get label => switch (this) {
        DeliveryStatus.created => 'Chờ xác nhận',
        DeliveryStatus.accepted => 'Đang đến điểm lấy hàng',
        DeliveryStatus.parcelPickedUp => 'Đã lấy hàng',
        DeliveryStatus.inDelivery => 'Đang giao hàng',
        DeliveryStatus.delivered => 'Đã giao hàng',
        DeliveryStatus.completed => 'Hoàn thành',
        DeliveryStatus.cancelled => 'Đã hủy',
        DeliveryStatus.unknown => 'Đang cập nhật',
      };
}
