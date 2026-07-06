import 'package:flutter/material.dart';

/// The five rider-facing trip lifecycle states covered by this module.
///
/// Mirrors the *shape* of the backend `TripStatus` state machine in
/// `backend/services/trip/domain/entity/trip.go` (pending → searching →
/// driver_assigned → driver_arrived → in_progress → completed) collapsed to
/// the states the rider UI actually needs to show. This is a UI-only mock
/// enum — no backend/API/proto code is referenced or generated from here.
enum RiderTripStatus {
  searchingDriver,
  driverAssigned,
  driverArriving,
  inProgress,
  completed,
}

extension RiderTripStatusX on RiderTripStatus {
  String get label => switch (this) {
        RiderTripStatus.searchingDriver => 'Searching Driver',
        RiderTripStatus.driverAssigned => 'Driver Assigned',
        RiderTripStatus.driverArriving => 'Driver Arriving',
        RiderTripStatus.inProgress => 'Trip In Progress',
        RiderTripStatus.completed => 'Trip Completed',
      };

  /// Short label used under the trip progress indicator.
  String get shortLabel => switch (this) {
        RiderTripStatus.searchingDriver => 'Search',
        RiderTripStatus.driverAssigned => 'Assigned',
        RiderTripStatus.driverArriving => 'Arriving',
        RiderTripStatus.inProgress => 'In trip',
        RiderTripStatus.completed => 'Done',
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
      };

  IconData get icon => switch (this) {
        RiderTripStatus.searchingDriver => Icons.search,
        RiderTripStatus.driverAssigned => Icons.person_pin_circle,
        RiderTripStatus.driverArriving => Icons.directions_car,
        RiderTripStatus.inProgress => Icons.route,
        RiderTripStatus.completed => Icons.check_circle,
      };

  /// Fraction of the trip progress indicator to fill. Mock/illustrative only.
  double get progressValue => switch (this) {
        RiderTripStatus.searchingDriver => 0.1,
        RiderTripStatus.driverAssigned => 0.35,
        RiderTripStatus.driverArriving => 0.55,
        RiderTripStatus.inProgress => 0.8,
        RiderTripStatus.completed => 1.0,
      };

  /// Cancel Ride is only offered before the trip has actually started.
  bool get isCancellable =>
      this == RiderTripStatus.searchingDriver ||
      this == RiderTripStatus.driverAssigned ||
      this == RiderTripStatus.driverArriving;

  /// No driver has been matched yet while searching.
  bool get hasDriver => this != RiderTripStatus.searchingDriver;

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
}
