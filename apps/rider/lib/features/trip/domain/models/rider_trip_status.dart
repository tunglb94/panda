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
        RiderTripStatus.searchingDriver => 'Searching Driver',
        RiderTripStatus.driverAssigned => 'Driver Assigned',
        RiderTripStatus.driverArriving => 'Driver Arriving',
        RiderTripStatus.inProgress => 'Trip In Progress',
        RiderTripStatus.completed => 'Trip Completed',
        RiderTripStatus.cancelled => 'Trip Cancelled',
        RiderTripStatus.paymentPending => 'Payment Pending',
        RiderTripStatus.paymentSuccess => 'Payment Successful',
        RiderTripStatus.settled => 'Trip Settled',
      };

  /// Short label used under the trip progress indicator.
  String get shortLabel => switch (this) {
        RiderTripStatus.searchingDriver => 'Search',
        RiderTripStatus.driverAssigned => 'Assigned',
        RiderTripStatus.driverArriving => 'Arriving',
        RiderTripStatus.inProgress => 'In trip',
        RiderTripStatus.completed => 'Done',
        RiderTripStatus.cancelled => 'Cancelled',
        RiderTripStatus.paymentPending => 'Pay',
        RiderTripStatus.paymentSuccess => 'Paid',
        RiderTripStatus.settled => 'Settled',
      };

  String get statusMessage => switch (this) {
        RiderTripStatus.searchingDriver =>
          'Looking for a nearby driver for you…',
        RiderTripStatus.driverAssigned =>
          'A driver has been assigned to your trip.',
        RiderTripStatus.driverArriving =>
          'Your driver is arriving at the pickup point.',
        RiderTripStatus.inProgress => 'You are on your way to the destination.',
        RiderTripStatus.completed => 'You have arrived. Trip completed.',
        RiderTripStatus.cancelled => 'Your trip was cancelled.',
        RiderTripStatus.paymentPending => 'Please pay to complete your trip.',
        RiderTripStatus.paymentSuccess => 'Payment received. Settling…',
        RiderTripStatus.settled => 'Trip settled. Thank you!',
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
