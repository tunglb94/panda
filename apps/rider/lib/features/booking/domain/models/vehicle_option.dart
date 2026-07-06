import 'package:flutter/material.dart';

/// Vehicle categories offered at MVP (per DOC-0002 §6.19 — Economy tier
/// only: car, motorcycle, van).
enum VehicleCategory { car, motorcycle, van }

/// A single selectable ride option shown in the Vehicle Selector.
///
/// Fare fields mirror the shape of `DefaultFareConfig` in
/// `backend/services/pricing/domain/entity/fare.go` (values in smallest
/// currency unit, e.g. USD cents) so the mock fare preview lands in the same
/// ballpark as what the real Pricing service will eventually return. This is
/// a UI-only mock model — no network call is made and no backend code is
/// referenced.
class VehicleOption {
  const VehicleOption({
    required this.category,
    required this.label,
    required this.icon,
    required this.capacity,
    required this.baseFareCents,
    required this.perKmCents,
    required this.perMinuteCents,
    required this.minimumFareCents,
    required this.bookingFeeCents,
  });

  final VehicleCategory category;
  final String label;
  final IconData icon;
  final int capacity;
  final int baseFareCents;
  final int perKmCents;
  final int perMinuteCents;
  final int minimumFareCents;
  final int bookingFeeCents;
}
