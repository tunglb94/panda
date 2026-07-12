import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/storage/trip_storage.dart';
import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_icon_sizes.dart';
import 'package:rider/core/theme/app_radius.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/booking/data/booking_repository.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/features/trip/presentation/pages/trip_lifecycle_page.dart';
import 'package:rider/shared/widgets/app_card.dart';

import '../../domain/models/mock_booking_catalog.dart';
import '../../domain/models/mock_fare_calculator.dart';
import '../../domain/models/mock_trip_metrics.dart';
import '../../domain/models/payment_method.dart';
import '../../domain/models/vehicle_option.dart';
import '../../domain/models/voucher.dart';
import 'book_ride_button.dart';
import 'fare_summary_card.dart';
import 'payment_method_card.dart';
import 'trip_point_cards.dart';
import 'vehicle_list_sheet.dart';
import 'voucher_list_sheet.dart';

/// Composes the full booking configuration form: trip summary, vehicle
/// choice, fare preview, payment method, voucher picker, and the Book Ride CTA.
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
  Voucher? _selectedVoucher;
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
        discountPercent: _selectedVoucher?.discountPercent ?? 0,
      );

  Future<void> _pickVoucher() async {
    final result = await VoucherListSheet.show(context, selected: _selectedVoucher);
    if (!mounted) return;
    setState(() => _selectedVoucher = result);
  }

  Future<void> _pickVehicle() async {
    final result = await VehicleListSheet.show(
      context,
      options: MockBookingCatalog.vehicles,
      selected: _selectedCategory,
      distanceKm: _distanceKm,
      durationMin: _durationMin,
    );
    if (!mounted || result == null) return;
    setState(() => _selectedCategory = result);
  }

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
      // statusCode 0 is only ever thrown client-side by ApiClient itself
      // (timeout/connectivity) with copy that's already Vietnamese and
      // safe to show verbatim; any real HTTP status is a raw backend
      // message and must never reach the rider as-is (see the same rule
      // applied in trip_lifecycle_page.dart's _pay/_poll).
      if (mounted) {
        setState(() => _bookingError = e.statusCode == 0 ? e.message : 'Đặt xe thất bại. Vui lòng thử lại.');
      }
      rethrow; // signal failure to BookRideButton so it resets to idle
    } catch (_) {
      if (mounted) {
        setState(() => _bookingError = 'Đặt xe thất bại. Vui lòng thử lại.');
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
        const RouteConnector(),
        DestinationCard(
          address: trip.destinationAddress,
          coordinate: trip.destination,
        ),
        const SizedBox(height: AppSpacing.xl),
        Text(
          'Chọn loại xe',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: AppSpacing.md),
        _VehicleSummaryRow(vehicle: _selectedVehicle, fare: fare, onTap: _pickVehicle),
        const SizedBox(height: AppSpacing.lg),
        AnimatedSwitcher(
          duration: const Duration(milliseconds: 250),
          child: FareSummaryCard(
            key: ValueKey('$_selectedCategory-${_selectedVoucher?.id}'),
            breakdown: fare,
            distanceKm: _distanceKm,
            durationMin: _durationMin,
            voucher: _selectedVoucher,
            // No promotion or surge data source exists anywhere in the
            // backend today (see PromotionInfo/SurgeInfo doc comments) —
            // both stay null rather than being fabricated.
            promotion: null,
            surge: null,
            cheaperThanCompetitorLabel: null,
          ),
        ),
        const SizedBox(height: AppSpacing.lg),
        PaymentMethodCard(
          selected: _selectedPayment,
          onChanged: (m) => setState(() => _selectedPayment = m),
        ),
        const SizedBox(height: AppSpacing.lg),
        VoucherPickerTile(selected: _selectedVoucher, onTap: _pickVoucher),
        if (_bookingError != null) ...[
          const SizedBox(height: AppSpacing.md),
          Text(
            _bookingError!,
            style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.error),
            textAlign: TextAlign.center,
          ),
        ],
        const SizedBox(height: AppSpacing.xl),
        SizedBox(
          width: double.infinity,
          child: BookRideButton(
            label: 'Đặt ${_selectedVehicle.label} · ${fare.format(fare.totalCents)}',
            onConfirm: _handleBookRide,
          ),
        ),
      ],
    );
  }
}

/// Tappable summary row shown in the booking form for the currently
/// selected vehicle — opens [VehicleListSheet] on tap. Replaces the old
/// always-visible horizontal-scroll picker with a Be/Xanh SM-style
/// "tap to open the full list" interaction.
class _VehicleSummaryRow extends StatelessWidget {
  const _VehicleSummaryRow({required this.vehicle, required this.fare, required this.onTap});

  final VehicleOption vehicle;
  final MockFareBreakdown fare;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    return AppCard(
      animateIn: false,
      onTap: onTap,
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
      child: Row(
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(color: AppColors.surfaceAlt, borderRadius: AppRadius.mdAll),
            child: Icon(vehicle.icon, color: AppColors.textPrimary, size: AppIconSize.lg),
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(vehicle.label, style: textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w700)),
                Text(fare.format(fare.totalCents), style: textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary)),
              ],
            ),
          ),
          Icon(Icons.chevron_right, size: AppIconSize.md, color: AppColors.textTertiary),
        ],
      ),
    );
  }
}
