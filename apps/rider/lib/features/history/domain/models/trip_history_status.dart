import 'package:flutter/material.dart';

/// Terminal outcome of a past trip. Distinct from `RiderTripStatus` (the
/// live 5-phase trip lifecycle in `features/trip`) — history only ever
/// shows a trip that has already ended one of these two ways.
enum TripHistoryStatus { completed, cancelled }

extension TripHistoryStatusX on TripHistoryStatus {
  String get label => switch (this) {
        TripHistoryStatus.completed => 'Completed',
        TripHistoryStatus.cancelled => 'Cancelled',
      };

  Color get color => switch (this) {
        TripHistoryStatus.completed => const Color(0xFF1A8C4E),
        TripHistoryStatus.cancelled => const Color(0xFFB91C1C),
      };

  IconData get icon => switch (this) {
        TripHistoryStatus.completed => Icons.check_circle,
        TripHistoryStatus.cancelled => Icons.cancel,
      };
}
