import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/features/map/domain/models/trip_selection.dart';

import 'payment_method.dart';
import 'vehicle_option.dart';

/// Central mock data source for the Booking UI module.
///
/// Every value here is a placeholder pending real backend wiring:
/// - vehicle rates mirror `backend/services/pricing` `DefaultFareConfig`
///   (Pricing service already exists — real wiring is Roadmap stage R4)
/// - payment methods stand in for the not-yet-started Wallet/Payment
///   services (Roadmap stage R6)
/// - [sampleTripSelection] stands in for a real `TripSelection` when the
///   Booking tab is opened directly from the bottom nav, without coming
///   from the Map's pickup/destination flow first
class MockBookingCatalog {
  const MockBookingCatalog._();

  static const List<VehicleOption> vehicles = [
    VehicleOption(
      category: VehicleCategory.car,
      label: 'Car',
      icon: Icons.directions_car,
      capacity: 4,
      baseFareCents: 50,
      perKmCents: 30,
      perMinuteCents: 5,
      minimumFareCents: 200,
      bookingFeeCents: 50,
    ),
    VehicleOption(
      category: VehicleCategory.motorcycle,
      label: 'Moto',
      icon: Icons.two_wheeler,
      capacity: 1,
      baseFareCents: 30,
      perKmCents: 20,
      perMinuteCents: 3,
      minimumFareCents: 150,
      bookingFeeCents: 30,
    ),
    VehicleOption(
      category: VehicleCategory.van,
      label: 'Van',
      icon: Icons.airport_shuttle,
      capacity: 6,
      baseFareCents: 100,
      perKmCents: 50,
      perMinuteCents: 8,
      minimumFareCents: 300,
      bookingFeeCents: 75,
    ),
  ];

  static const List<PaymentMethod> paymentMethods = [
    PaymentMethod(
      type: PaymentMethodType.cash,
      label: 'Cash',
      subtitle: 'Pay the driver directly',
      icon: Icons.payments_outlined,
    ),
    PaymentMethod(
      type: PaymentMethodType.wallet,
      label: 'FAIRRIDE Wallet',
      subtitle: 'Balance: \$12.40 (mock)',
      icon: Icons.account_balance_wallet_outlined,
    ),
    PaymentMethod(
      type: PaymentMethodType.card,
      label: 'Visa •••• 4242',
      subtitle: 'Expires 08/28',
      icon: Icons.credit_card,
    ),
  ];

  /// Used only when the Booking tab is opened directly (bottom nav) with no
  /// [TripSelection] carried over from the Map page's pickup/destination
  /// flow, so the screen is always demoable end-to-end.
  static const TripSelection sampleTripSelection = TripSelection(
    pickup: LatLng(10.7769, 106.7009),
    destination: LatLng(10.8231, 106.6297),
    pickupAddress: 'Current location (sample)',
    destinationAddress: 'Sample destination',
  );
}
