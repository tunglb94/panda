import 'package:flutter/material.dart';

import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/features/trip/presentation/pages/trip_lifecycle_page.dart';

import '../../domain/models/mock_booking_catalog.dart';
import '../../domain/models/mock_fare_calculator.dart';
import '../../domain/models/mock_trip_metrics.dart';
import '../../domain/models/payment_method.dart';
import '../../domain/models/promo_result.dart';
import '../../domain/models/vehicle_option.dart';
import 'book_ride_button.dart';
import 'fare_summary_card.dart';
import 'payment_method_card.dart';
import 'promo_code_entry.dart';
import 'trip_point_cards.dart';
import 'vehicle_selector.dart';

/// Composes the full booking configuration form: trip summary, vehicle
/// choice, fare preview, payment method, promo code, and the Book Ride CTA.
///
/// Reused by both `BookingPage` (full-page, bottom-nav entry) and
/// `BookingBottomSheet` (modal, invoked from the Map's confirmed selection)
/// so both entry points share one implementation. All data is mock — see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.1/§2.2 for what still needs
/// real backend wiring (Roadmap stages R2/R4/R6).
class BookingFormBody extends StatefulWidget {
  const BookingFormBody({super.key, required this.tripSelection});

  final TripSelection tripSelection;

  @override
  State<BookingFormBody> createState() => _BookingFormBodyState();
}

class _BookingFormBodyState extends State<BookingFormBody> {
  VehicleCategory _selectedCategory = VehicleCategory.car;
  PaymentMethod _selectedPayment = MockBookingCatalog.paymentMethods.first;
  PromoResult _promo = PromoResult.none;

  late final double _distanceKm;
  late final double _durationMin;

  @override
  void initState() {
    super.initState();
    _distanceKm = MockTripMetrics.distanceKm(
      widget.tripSelection.pickup,
      widget.tripSelection.destination,
    );
    _durationMin = MockTripMetrics.estimateDurationMinutes(_distanceKm);
  }

  VehicleOption get _selectedVehicle => MockBookingCatalog.vehicles
      .firstWhere((v) => v.category == _selectedCategory);

  MockFareBreakdown get _fare => MockFareBreakdown.calculate(
        vehicle: _selectedVehicle,
        distanceKm: _distanceKm,
        durationMin: _durationMin,
        discountPercent: _promo.discountPercent,
      );

  Future<void> _handleBookRide() async {
    // Mock submission only — no API call. Real submission wiring is
    // MVP_DEVELOPMENT_PLAN.md Rider App Roadmap stage R4.
    final fareAtBooking = _fare;
    await Future.delayed(const Duration(milliseconds: 1200));
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text(
          'Mock ride requested — tracking your trip (UI only, not wired to '
          'the backend; see MVP roadmap R4).',
        ),
      ),
    );
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => TripLifecyclePage(
          tripSelection: widget.tripSelection,
          fare: fareAtBooking,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final trip = widget.tripSelection;
    final fare = _fare;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        PickupCard(address: trip.pickupAddress, coordinate: trip.pickup),
        const SizedBox(height: 8),
        DestinationCard(
          address: trip.destinationAddress,
          coordinate: trip.destination,
        ),
        const SizedBox(height: 20),
        Text(
          'Choose a ride',
          style: Theme.of(context)
              .textTheme
              .titleMedium
              ?.copyWith(fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 12),
        VehicleSelector(
          options: MockBookingCatalog.vehicles,
          selected: _selectedCategory,
          distanceKm: _distanceKm,
          durationMin: _durationMin,
          onSelected: (category) => setState(() => _selectedCategory = category),
        ),
        const SizedBox(height: 16),
        AnimatedSwitcher(
          duration: const Duration(milliseconds: 250),
          child: FareSummaryCard(
            key: ValueKey('$_selectedCategory-${_promo.discountPercent}'),
            breakdown: fare,
          ),
        ),
        const SizedBox(height: 16),
        PaymentMethodCard(
          selected: _selectedPayment,
          onChanged: (m) => setState(() => _selectedPayment = m),
        ),
        const SizedBox(height: 16),
        PromoCodeEntry(
          onApplied: (result) => setState(() => _promo = result),
        ),
        const SizedBox(height: 20),
        SizedBox(
          width: double.infinity,
          child: BookRideButton(
            label: 'Book ${_selectedVehicle.label} · ${fare.format(fare.totalCents)}',
            onConfirm: _handleBookRide,
          ),
        ),
      ],
    );
  }
}
