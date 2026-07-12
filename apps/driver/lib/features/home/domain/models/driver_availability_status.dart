import 'package:flutter/material.dart';

/// Visual/animation states of the big availability toggle on the Home
/// dashboard. `goingOnline`/`goingOffline` are transient — the toggle is
/// disabled while in either, driven by a local mock delay (see
/// `AvailabilityToggle`), not a real background-location/API call.
enum DriverAvailabilityStatus { offline, goingOnline, online, goingOffline }

extension DriverAvailabilityStatusX on DriverAvailabilityStatus {
  String get label => switch (this) {
        DriverAvailabilityStatus.offline => 'Ngoại tuyến',
        DriverAvailabilityStatus.goingOnline => 'Đang bật trực tuyến',
        DriverAvailabilityStatus.online => 'Trực tuyến',
        DriverAvailabilityStatus.goingOffline => 'Đang tắt trực tuyến',
      };

  /// Prompt shown on the toggle itself.
  String get actionLabel => switch (this) {
        DriverAvailabilityStatus.offline => "Bạn đang ngoại tuyến — chạm để bật trực tuyến",
        DriverAvailabilityStatus.goingOnline => 'Đang bật trực tuyến…',
        DriverAvailabilityStatus.online => "Bạn đang trực tuyến — chạm để tắt trực tuyến",
        DriverAvailabilityStatus.goingOffline => 'Đang tắt trực tuyến…',
      };

  IconData get icon => switch (this) {
        DriverAvailabilityStatus.offline => Icons.power_settings_new,
        DriverAvailabilityStatus.goingOnline => Icons.sync,
        DriverAvailabilityStatus.online => Icons.bolt,
        DriverAvailabilityStatus.goingOffline => Icons.sync,
      };

  bool get isTransitioning =>
      this == DriverAvailabilityStatus.goingOnline ||
      this == DriverAvailabilityStatus.goingOffline;

  bool get isOnlineOrBecomingOnline =>
      this == DriverAvailabilityStatus.online ||
      this == DriverAvailabilityStatus.goingOnline;
}
