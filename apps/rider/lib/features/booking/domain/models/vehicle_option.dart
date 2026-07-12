import 'package:flutter/material.dart';

/// Ride tiers in Panda's Vehicle Catalog (backend Vehicle Catalog
/// Expansion — see `backend/services/pricing/domain/entity/fare.go`).
/// `motorcycle`/`car` are the two original tiers (product-facing "Bike"/
/// "Car"); `bikePlus`/`carXL` are the two new tiers added alongside them.
/// `van` (backend's 3rd original value) is legacy — not part of the
/// rider-facing Ride catalog, so it has no entry here.
enum VehicleCategory { motorcycle, bikePlus, car, carXL }

/// The vehicle "family" a [VehicleCategory] belongs to — used to group tabs
/// in the vehicle-list sheet (Xe máy vs Ô tô).
enum VehicleFamily { bike, car }

extension VehicleCategoryX on VehicleCategory {
  /// The exact string key the backend's VehicleType recognizes
  /// (`backend/services/pricing/domain/entity/fare.go`,
  /// `backend/services/dispatch/domain/entity/dispatch_job.go`).
  String get backendKey => switch (this) {
        VehicleCategory.motorcycle => 'motorcycle',
        VehicleCategory.bikePlus => 'bike_plus',
        VehicleCategory.car => 'car',
        VehicleCategory.carXL => 'car_xl',
      };

  VehicleFamily get family => switch (this) {
        VehicleCategory.motorcycle || VehicleCategory.bikePlus => VehicleFamily.bike,
        VehicleCategory.car || VehicleCategory.carXL => VehicleFamily.car,
      };
}

/// A single selectable ride option shown in the Vehicle Selector.
///
/// Fare fields mirror the shape and values of `DefaultFareConfig` in
/// `backend/services/pricing/domain/entity/fare.go` (whole VND — no decimal
/// subunit — despite the legacy `*Cents` field names) so the mock fare
/// preview matches what the real Pricing service returns. This is a UI-only
/// mock model — no network call is made and no backend code is referenced.
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
    this.isAvailable = true,
    this.originalPriceCents,
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

  /// False for tiers the backend recognizes but has no BRB-approved fare
  /// config for yet (Bike Plus, Car XL — see
  /// `backend/services/pricing/config/pricing_v3.default.yaml`'s
  /// placeholder comment). Shown in the list as "Chưa khả dụng", not
  /// selectable, no price rendered — never a fabricated number.
  final bool isAvailable;

  /// The pre-discount price, shown struck through next to the real price —
  /// ONLY when a genuine discount exists (e.g. an auto-applied promotion).
  /// Null means no discount; the UI must never invent this value. Nothing
  /// in the backend today auto-applies a promotion before vehicle
  /// selection, so this is always null until that exists — see the
  /// Vehicle Catalog Expansion frontend report's "Known Gap" section.
  final int? originalPriceCents;
}
