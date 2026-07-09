import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/storage/trip_storage.dart';
import 'package:rider/features/booking/data/booking_repository.dart';
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
/// `BookingBottomSheet` (modal, invoked from the Map's confirmed selection).
class BookingFormBody extends StatefulWidget {
  const BookingFormBody({
    super.key,
    required this.tripSelection,
    required this.apiClient,
    this.onDriverAssigned,
  });

  final TripSelection tripSelection;
  final ApiClient apiClient;
  final void Function(String driverId)? onDriverAssigned;

  @override
  State<BookingFormBody> createState() => _BookingFormBodyState();
}

class _BookingFormBodyState extends State<BookingFormBody> {
  VehicleCategory _selectedCategory = VehicleCategory.car;
  PaymentMethod _selectedPayment = MockBookingCatalog.paymentMethods.first;
  PromoResult _promo = PromoResult.none;
  String? _bookingError;

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
    setState(() => _bookingError = null);

    try {
      final repo = BookingRepository(widget.apiClient);
      final result = await repo.bookRide(widget.tripSelection);

      await TripStorage().saveActiveTripId(result.tripId);

      if (!mounted) return;
      await Navigator.of(context).push(
        MaterialPageRoute(
          builder: (_) => TripLifecyclePage(
            tripId: result.tripId,
            tripSelection: widget.tripSelection,
            apiClient: widget.apiClient,
            onDriverAssigned: widget.onDriverAssigned,
          ),
        ),
      );
      // Trip ended (completed or cancelled) — clear the persisted trip.
      await TripStorage().clearActiveTripId();
    } on ApiException catch (e) {
      if (mounted) setState(() => _bookingError = e.message);
      rethrow; // signal failure to BookRideButton so it resets to idle
    } catch (_) {
      if (mounted) {
        setState(() => _bookingError = 'Booking failed. Please try again.');
      }
      rethrow;
    }
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
        if (_bookingError != null) ...[
          const SizedBox(height: 12),
          Text(
            _bookingError!,
            style: TextStyle(
                color: Theme.of(context).colorScheme.error, fontSize: 13),
            textAlign: TextAlign.center,
          ),
        ],
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
