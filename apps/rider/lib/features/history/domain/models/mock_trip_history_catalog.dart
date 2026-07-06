import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/features/booking/domain/models/mock_fare_calculator.dart';
import 'package:rider/features/booking/domain/models/payment_method.dart';
import 'package:rider/features/booking/domain/models/vehicle_option.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/features/trip/domain/models/mock_driver.dart';

import 'trip_history_entry.dart';
import 'trip_history_status.dart';
import 'trip_timeline_event.dart';

/// Sample trip history, spanning today / this week / this month / older —
/// so the date-range filters (Today / This Week / This Month) each have at
/// least one matching and one non-matching entry to filter against.
class MockTripHistoryCatalog {
  const MockTripHistoryCatalog._();

  static List<TripHistoryEntry> sample() {
    final now = DateTime.now();
    return [
      _entry(
        id: 'trip-1',
        when: now.subtract(const Duration(hours: 2)),
        pickup: 'Current location',
        destination: 'FAIRRIDE Office',
        status: TripHistoryStatus.completed,
        driverName: 'Nguyen Van A',
        vehicleModel: 'Toyota Vios',
        plate: '51G-123.45',
        category: VehicleCategory.car,
        distanceKm: 6.4,
        durationMin: 14,
        ratingGiven: 5,
        payment: _cash,
      ),
      _entry(
        id: 'trip-2',
        when: now.subtract(const Duration(days: 2, hours: 3)),
        pickup: 'District 1 Market',
        destination: 'Tan Son Nhat Airport',
        status: TripHistoryStatus.completed,
        driverName: 'Tran Thi B',
        vehicleModel: 'Honda Wave',
        plate: '59H1-678.90',
        category: VehicleCategory.motorcycle,
        distanceKm: 11.2,
        durationMin: 26,
        ratingGiven: 4.5,
        payment: _wallet,
        discountPercent: 10,
      ),
      _entry(
        id: 'trip-3',
        when: now.subtract(const Duration(days: 5)),
        pickup: 'Ben Thanh Market',
        destination: 'Landmark 81',
        status: TripHistoryStatus.cancelled,
        driverName: 'Le Van C',
        vehicleModel: 'Ford Transit',
        plate: '51F-222.33',
        category: VehicleCategory.van,
        distanceKm: 8.9,
        durationMin: 0,
        ratingGiven: null,
        payment: _card,
        cancellationReason: 'Rider cancelled before driver arrived',
      ),
      _entry(
        id: 'trip-4',
        when: now.subtract(const Duration(days: 12)),
        pickup: 'Home',
        destination: 'District 7 Mall',
        status: TripHistoryStatus.completed,
        driverName: 'Pham Thi D',
        vehicleModel: 'Toyota Innova',
        plate: '51G-999.11',
        category: VehicleCategory.car,
        distanceKm: 14.6,
        durationMin: 32,
        ratingGiven: 4,
        payment: _cash,
      ),
      _entry(
        id: 'trip-5',
        when: now.subtract(const Duration(days: 45)),
        pickup: 'Home',
        destination: 'Vung Tau Bus Station',
        status: TripHistoryStatus.completed,
        driverName: 'Hoang Van E',
        vehicleModel: 'Hyundai Grand i10',
        plate: '72A-456.78',
        category: VehicleCategory.car,
        distanceKm: 22.1,
        durationMin: 41,
        ratingGiven: 5,
        payment: _wallet,
      ),
    ];
  }

  static const _cash = PaymentMethod(
    type: PaymentMethodType.cash,
    label: 'Cash',
    subtitle: 'Paid the driver directly',
    icon: Icons.payments_outlined,
  );

  static const _wallet = PaymentMethod(
    type: PaymentMethodType.wallet,
    label: 'FAIRRIDE Wallet',
    subtitle: 'Paid from wallet balance',
    icon: Icons.account_balance_wallet_outlined,
  );

  static const _card = PaymentMethod(
    type: PaymentMethodType.card,
    label: 'Visa •••• 4242',
    subtitle: 'Charged to card',
    icon: Icons.credit_card,
  );

  static TripHistoryEntry _entry({
    required String id,
    required DateTime when,
    required String pickup,
    required String destination,
    required TripHistoryStatus status,
    required String driverName,
    required String vehicleModel,
    required String plate,
    required VehicleCategory category,
    required double distanceKm,
    required double durationMin,
    required double? ratingGiven,
    required PaymentMethod payment,
    int discountPercent = 0,
    String? cancellationReason,
  }) {
    final vehicle = _ratesFor(category);
    final fare = MockFareBreakdown.calculate(
      vehicle: vehicle,
      distanceKm: distanceKm,
      durationMin: durationMin,
      discountPercent: discountPercent,
    );

    final timeline = status == TripHistoryStatus.cancelled
        ? [
            TripTimelineEvent(label: 'Requested', time: when, icon: Icons.search),
            TripTimelineEvent(
              label: 'Cancelled',
              time: when.add(const Duration(minutes: 2)),
              icon: Icons.cancel,
            ),
          ]
        : [
            TripTimelineEvent(label: 'Requested', time: when, icon: Icons.search),
            TripTimelineEvent(
              label: 'Driver assigned',
              time: when.add(const Duration(minutes: 1)),
              icon: Icons.person_pin_circle,
            ),
            TripTimelineEvent(
              label: 'Trip started',
              time: when.add(const Duration(minutes: 4)),
              icon: Icons.route,
            ),
            TripTimelineEvent(
              label: 'Trip completed',
              time: when.add(Duration(minutes: 4 + durationMin.round())),
              icon: Icons.check_circle,
            ),
          ];

    return TripHistoryEntry(
      id: id,
      dateTime: when,
      route: TripSelection(
        // Sample coordinates only (Ho Chi Minh City area) — addresses are
        // what the UI actually displays.
        pickup: const LatLng(10.7769, 106.7009),
        destination: const LatLng(10.8231, 106.6297),
        pickupAddress: pickup,
        destinationAddress: destination,
      ),
      status: status,
      driver: MockDriver(
        name: driverName,
        vehicleModel: vehicleModel,
        plateNumber: plate,
        rating: 4.8,
      ),
      vehicleCategory: category,
      fare: fare,
      paymentMethod: payment,
      distanceKm: distanceKm,
      durationMin: durationMin,
      timeline: timeline,
      ratingGiven: ratingGiven,
      cancellationReason: cancellationReason,
    );
  }

  /// Same default rates as `MockBookingCatalog.vehicles` — kept as a local
  /// copy so the history module doesn't reach into the Booking catalog's
  /// full `VehicleOption` list just for rate constants.
  static VehicleOption _ratesFor(VehicleCategory category) => switch (category) {
        VehicleCategory.car => const VehicleOption(
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
        VehicleCategory.motorcycle => const VehicleOption(
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
        VehicleCategory.van => const VehicleOption(
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
      };
}
