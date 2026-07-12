import 'package:flutter/material.dart';

/// Message state of the Home Status Card. An independent (finer-grained)
/// axis from `DriverAvailabilityStatus`: once truly online, the driver is
/// either waiting, actively being searched for, or (mock/placeholder) busy
/// on a trip. There is no real trip/dispatch system wired in yet (Driver
/// App Roadmap stages D4/D6), so [busy] is only reachable here via the Home
/// page's dev "Preview state" menu — it is never entered through the
/// natural offline/online flow in this phase.
enum DriverActivityStatus { offline, waitingForTrips, searchingNearby, busy }

extension DriverActivityStatusX on DriverActivityStatus {
  String get title => switch (this) {
        DriverActivityStatus.offline => "Bạn đang ngoại tuyến",
        DriverActivityStatus.waitingForTrips => 'Đang chờ chuyến',
        DriverActivityStatus.searchingNearby => 'Đang tìm gần đây',
        DriverActivityStatus.busy => 'Đang trong chuyến (placeholder)',
      };

  String get message => switch (this) {
        DriverActivityStatus.offline =>
          'Bật trực tuyến để bắt đầu nhận yêu cầu chuyến đi.',
        DriverActivityStatus.waitingForTrips =>
          "Bạn đang trực tuyến. Chờ chút — yêu cầu chuyến sẽ hiện ở đây.",
        DriverActivityStatus.searchingNearby =>
          'Đang kiểm tra xung quanh tìm hành khách cần đi xe.',
        DriverActivityStatus.busy =>
          'Đây là placeholder — việc gán chuyến đi chưa được kết nối '
              '(xem Driver App Roadmap stages D4/D6).',
      };

  IconData get icon => switch (this) {
        DriverActivityStatus.offline => Icons.power_settings_new,
        DriverActivityStatus.waitingForTrips => Icons.hourglass_empty,
        DriverActivityStatus.searchingNearby => Icons.search,
        DriverActivityStatus.busy => Icons.local_taxi,
      };
}
