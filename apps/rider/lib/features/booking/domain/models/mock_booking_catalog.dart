import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/features/map/domain/models/trip_selection.dart';

import 'payment_method.dart';
import 'vehicle_option.dart';

/// Central mock data source for the Booking UI module.
///
/// Every value here is a placeholder pending real backend wiring:
/// - vehicle rates mirror `backend/services/pricing` `DefaultFareConfig`,
///   i.e. Business Rule Bible v1.0 §2.2.1-§2.2.5 (Standard→car, XL→van;
///   BRB v1.0 defines no motorcycle rate, so those figures are an interim
///   estimate only — see the comment on `DefaultFareConfig` in
///   `backend/services/pricing/domain/entity/fare.go`). Pricing service
///   already exists — real network wiring is Roadmap stage R4.
/// - payment methods stand in for the not-yet-started Wallet/Payment
///   services (Roadmap stage R6)
/// - [sampleTripSelection] stands in for a real `TripSelection` when the
///   Booking tab is opened directly from the bottom nav, without coming
///   from the Map's pickup/destination flow first
class MockBookingCatalog {
  const MockBookingCatalog._();

  // Vehicle Catalog Expansion (backend): Bike (motorcycle) and Car keep
  // their original BRB-derived rates below. Bike Plus / Car XL are
  // recognized by the backend's VehicleType allow-list but have no
  // BRB-approved fare config yet (pricing_v3.default.yaml's placeholder
  // comment) — shown here with isAvailable: false and zero rate fields
  // (never read/displayed) rather than an invented price.
  static const List<VehicleOption> vehicles = [
    VehicleOption(
      category: VehicleCategory.motorcycle,
      label: 'Bike',
      icon: Icons.two_wheeler,
      capacity: 1,
      baseFareCents: 5000,
      perKmCents: 1600,
      perMinuteCents: 200,
      minimumFareCents: 12000,
      bookingFeeCents: 2000,
    ),
    VehicleOption(
      category: VehicleCategory.bikePlus,
      label: 'Bike Plus',
      icon: Icons.two_wheeler,
      capacity: 1,
      baseFareCents: 0,
      perKmCents: 0,
      perMinuteCents: 0,
      minimumFareCents: 0,
      bookingFeeCents: 0,
      isAvailable: false,
    ),
    VehicleOption(
      category: VehicleCategory.car,
      label: 'Car',
      icon: Icons.directions_car,
      capacity: 4,
      baseFareCents: 10000,
      perKmCents: 4000,
      perMinuteCents: 400,
      minimumFareCents: 25000,
      bookingFeeCents: 2000,
    ),
    VehicleOption(
      category: VehicleCategory.carXL,
      label: 'Car XL',
      icon: Icons.airport_shuttle,
      capacity: 7,
      baseFareCents: 0,
      perKmCents: 0,
      perMinuteCents: 0,
      minimumFareCents: 0,
      bookingFeeCents: 0,
      isAvailable: false,
    ),
  ];

  static const List<PaymentMethod> paymentMethods = [
    PaymentMethod(
      type: PaymentMethodType.cash,
      label: 'Tiền mặt',
      subtitle: 'Thanh toán trực tiếp cho tài xế',
      icon: Icons.payments_outlined,
    ),
    PaymentMethod(
      type: PaymentMethodType.wallet,
      label: 'Ví Panda',
      subtitle: 'Số dư: 310.000 đ (giả lập)',
      icon: Icons.account_balance_wallet_outlined,
    ),
    PaymentMethod(
      type: PaymentMethodType.card,
      label: 'Visa •••• 4242',
      subtitle: 'Hết hạn 08/28',
      icon: Icons.credit_card,
    ),
  ];

  /// Used only when the Booking tab is opened directly (bottom nav) with no
  /// [TripSelection] carried over from the Map page's pickup/destination
  /// flow, so the screen is always demoable end-to-end.
  static const TripSelection sampleTripSelection = TripSelection(
    pickup: LatLng(10.7769, 106.7009),
    destination: LatLng(10.8231, 106.6297),
    pickupAddress: 'Vị trí hiện tại (mẫu)',
    destinationAddress: 'Điểm đến mẫu',
  );
}
