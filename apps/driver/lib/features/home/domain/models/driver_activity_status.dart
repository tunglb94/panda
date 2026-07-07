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
        DriverActivityStatus.offline => "You're offline",
        DriverActivityStatus.waitingForTrips => 'Waiting for trips',
        DriverActivityStatus.searchingNearby => 'Searching nearby',
        DriverActivityStatus.busy => 'On a trip (placeholder)',
      };

  String get message => switch (this) {
        DriverActivityStatus.offline =>
          'Go online to start receiving trip requests.',
        DriverActivityStatus.waitingForTrips =>
          "You're online. Sit tight — trip requests will appear here.",
        DriverActivityStatus.searchingNearby =>
          'Checking nearby for riders looking for a ride.',
        DriverActivityStatus.busy =>
          'This is a placeholder — trip assignment is not wired up yet '
              '(see Driver App Roadmap stages D4/D6).',
      };

  IconData get icon => switch (this) {
        DriverActivityStatus.offline => Icons.power_settings_new,
        DriverActivityStatus.waitingForTrips => Icons.hourglass_empty,
        DriverActivityStatus.searchingNearby => Icons.search,
        DriverActivityStatus.busy => Icons.local_taxi,
      };
}
