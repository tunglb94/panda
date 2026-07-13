import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/features/map/domain/models/trip_selection.dart';

import 'payment_method.dart';
import 'vehicle_option.dart';

/// Static product catalog data for the Booking UI module — **not** a pricing
/// mock. Fare calculation has been removed entirely from Flutter (see
/// `PricingRepository.estimateFare`); everything left here is legitimate
/// static config Panda controls directly:
/// - [vehicles]: the tier identities/artwork/capacity Panda offers (no fare
///   fields — those come from the backend for every quote).
/// - [paymentMethods]: stands in for the not-yet-started Wallet/Payment
///   services (Roadmap stage R6).
/// - [sampleTripSelection]: stands in for a real `TripSelection` when the
///   Booking tab is opened directly from the bottom nav, without coming
///   from the Map's pickup/destination flow first.
class MockBookingCatalog {
  const MockBookingCatalog._();

  static const List<VehicleOption> vehicles = [
    VehicleOption(
      category: VehicleCategory.motorcycle,
      label: 'Bike',
      icon: Icons.two_wheeler,
      imageAsset: 'assets/vehicles/bike.png',
      capacity: 1,
    ),
    VehicleOption(
      category: VehicleCategory.bikePlus,
      label: 'Bike Plus',
      icon: Icons.two_wheeler,
      imageAsset: 'assets/vehicles/bike_plus.png',
      capacity: 1,
    ),
    VehicleOption(
      category: VehicleCategory.car,
      label: 'Car',
      icon: Icons.directions_car,
      imageAsset: 'assets/vehicles/car.png',
      capacity: 4,
    ),
    VehicleOption(
      category: VehicleCategory.carXL,
      label: 'Car XL',
      icon: Icons.airport_shuttle,
      imageAsset: 'assets/vehicles/car_xl.png',
      capacity: 7,
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
