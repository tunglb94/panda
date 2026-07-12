import 'package:flutter/material.dart';

/// The rider-facing trip lifecycle states.
///
/// Maps to the backend `TripStatus` state machine in
/// `backend/services/trip/domain/entity/trip.go`. The `cancelled` state was
/// added in Phase 29 when real backend polling replaced the mock repository.
enum RiderTripStatus {
  searchingDriver,
  driverAssigned,
  driverArriving,
  inProgress,
  completed,
  cancelled,
  paymentPending,
  paymentSuccess,
  settled,
}

extension RiderTripStatusX on RiderTripStatus {
  String get label => switch (this) {
        RiderTripStatus.searchingDriver => 'Đang tìm tài xế',
        RiderTripStatus.driverAssigned => 'Đã tìm thấy tài xế',
        RiderTripStatus.driverArriving => 'Tài xế đang đến',
        RiderTripStatus.inProgress => 'Đang trong chuyến đi',
        RiderTripStatus.completed => 'Chuyến đi hoàn tất',
        RiderTripStatus.cancelled => 'Chuyến đi đã hủy',
        RiderTripStatus.paymentPending => 'Đang chờ thanh toán',
        RiderTripStatus.paymentSuccess => 'Thanh toán thành công',
        RiderTripStatus.settled => 'Hoàn tất chuyến đi',
      };

  /// Short label used under the trip progress indicator.
  String get shortLabel => switch (this) {
        RiderTripStatus.searchingDriver => 'Tìm xe',
        RiderTripStatus.driverAssigned => 'Đã ghép',
        RiderTripStatus.driverArriving => 'Đang đến',
        RiderTripStatus.inProgress => 'Đang đi',
        RiderTripStatus.completed => 'Xong',
        RiderTripStatus.cancelled => 'Đã hủy',
        RiderTripStatus.paymentPending => 'Thanh toán',
        RiderTripStatus.paymentSuccess => 'Đã trả',
        RiderTripStatus.settled => 'Hoàn tất',
      };

  String get statusMessage => switch (this) {
        RiderTripStatus.searchingDriver =>
          'Đang tìm tài xế gần bạn…',
        RiderTripStatus.driverAssigned =>
          'Đã tìm thấy tài xế cho chuyến đi của bạn.',
        RiderTripStatus.driverArriving =>
          'Tài xế đang đến điểm đón.',
        RiderTripStatus.inProgress => 'Bạn đang trên đường đến điểm đến.',
        RiderTripStatus.completed => 'Bạn đã đến nơi. Chuyến đi hoàn tất.',
        RiderTripStatus.cancelled => 'Chuyến đi của bạn đã bị hủy.',
        RiderTripStatus.paymentPending => 'Vui lòng thanh toán để hoàn tất chuyến đi.',
        RiderTripStatus.paymentSuccess => 'Đã nhận thanh toán. Đang xử lý…',
        RiderTripStatus.settled => 'Chuyến đi đã hoàn tất. Cảm ơn bạn!',
      };

  IconData get icon => switch (this) {
        RiderTripStatus.searchingDriver => Icons.search,
        RiderTripStatus.driverAssigned => Icons.person_pin_circle,
        RiderTripStatus.driverArriving => Icons.directions_car,
        RiderTripStatus.inProgress => Icons.route,
        RiderTripStatus.completed => Icons.check_circle,
        RiderTripStatus.cancelled => Icons.cancel_outlined,
        RiderTripStatus.paymentPending => Icons.payment,
        RiderTripStatus.paymentSuccess => Icons.check_circle_outline,
        RiderTripStatus.settled => Icons.verified,
      };

  /// Fraction of the trip progress indicator to fill.
  double get progressValue => switch (this) {
        RiderTripStatus.searchingDriver => 0.1,
        RiderTripStatus.driverAssigned => 0.35,
        RiderTripStatus.driverArriving => 0.55,
        RiderTripStatus.inProgress => 0.8,
        RiderTripStatus.completed => 0.9,
        RiderTripStatus.cancelled => 0.0,
        RiderTripStatus.paymentPending => 0.92,
        RiderTripStatus.paymentSuccess => 0.96,
        RiderTripStatus.settled => 1.0,
      };

  /// Cancel Ride is only offered before the trip has actually started.
  bool get isCancellable =>
      this == RiderTripStatus.searchingDriver ||
      this == RiderTripStatus.driverAssigned ||
      this == RiderTripStatus.driverArriving;

  bool get hasDriver => this != RiderTripStatus.searchingDriver &&
      this != RiderTripStatus.cancelled;

  /// ETA/arrival card is only meaningful while a driver is en route.
  bool get showsEta =>
      this == RiderTripStatus.driverAssigned ||
      this == RiderTripStatus.driverArriving ||
      this == RiderTripStatus.inProgress;

  /// Contact Driver / Emergency stay available for the whole active trip.
  bool get showsSafetyActions =>
      this == RiderTripStatus.driverAssigned ||
      this == RiderTripStatus.driverArriving ||
      this == RiderTripStatus.inProgress;

  bool get isTerminal =>
      this == RiderTripStatus.settled || this == RiderTripStatus.cancelled;
}
