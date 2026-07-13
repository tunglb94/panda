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
  /// The exact string key the backend recognizes as `service_type` — both
  /// `POST /api/v1/rides/estimate-fare` (gateway's `PricingHandler`, which
  /// maps this onto Pricing's real `car|motorcycle|van` VehicleType) and the
  /// existing `x-service-type` dispatch metadata use this same key.
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

/// A single selectable ride option shown in the Vehicle Selector — pure
/// product catalog metadata (tier identity, label, artwork, capacity). Fare
/// is never stored here: every price shown anywhere in the app comes from a
/// live call to `PricingRepository.estimateFare` (backend Pricing service),
/// never a field on this class.
class VehicleOption {
  const VehicleOption({
    required this.category,
    required this.label,
    required this.icon,
    required this.capacity,
    this.isAvailable = true,
    this.imageAsset,
  });

  final VehicleCategory category;
  final String label;
  final IconData icon;

  /// Real vehicle artwork (`assets/vehicles/...`), shown instead of [icon]
  /// when present. Null falls back to [icon] — keeps every category
  /// renderable even before/without dedicated art.
  final String? imageAsset;
  final int capacity;

  /// False for a tier that isn't bookable yet — shown in the list as "Chưa
  /// khả dụng", not selectable, and never quoted a fare.
  final bool isAvailable;
}
