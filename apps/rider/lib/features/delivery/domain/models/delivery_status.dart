/// Mirrors `backend/services/trip/domain/entity/delivery.go`'s
/// `DeliveryStatus` enum exactly — CREATED/ACCEPTED/PARCEL_PICKED_UP/
/// IN_DELIVERY/DELIVERED/COMPLETED/CANCELLED. This is intentionally a
/// separate state machine from `RiderTripStatus`: `Trip.Status` for a
/// delivery trip does not advance the same way a Ride's does (see
/// `CompleteDeliveryUseCase`'s doc comment — Trip.Status is deliberately
/// left at `in_progress` through delivery completion, since there is no
/// fare/settlement step in this phase), so the delivery lifecycle screen
/// must drive off `delivery_status`, not `trip_status`.
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
        DeliveryStatus.created => 'Đang tìm tài xế',
        DeliveryStatus.accepted => 'Tài xế đang đến điểm lấy hàng',
        DeliveryStatus.parcelPickedUp => 'Đã lấy hàng',
        DeliveryStatus.inDelivery => 'Đang giao hàng',
        DeliveryStatus.delivered => 'Đã giao hàng',
        DeliveryStatus.completed => 'Hoàn thành',
        DeliveryStatus.cancelled => 'Đã hủy',
        DeliveryStatus.unknown => 'Đang cập nhật',
      };
}
