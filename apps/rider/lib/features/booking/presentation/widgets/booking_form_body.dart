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
import 'package:rider/shared/utils/currency_format.dart';
import 'package:rider/shared/widgets/app_button.dart';
import 'package:rider/shared/widgets/app_card.dart';

import '../../data/pricing_repository.dart';
import '../../data/promotion_repository.dart';
import '../../domain/models/fare_estimate.dart';
import '../../domain/models/mock_booking_catalog.dart';
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
/// The fare shown is always the real backend quote from
/// `PricingRepository.estimateFare` (`POST /api/v1/rides/estimate-fare`) —
/// Flutter performs no fare math of its own. Backend is the single source
/// of truth: a failed quote shows an error, never a locally-computed
/// fallback number.
///
/// Reused by both `BookingPage` (full-page, bottom-nav entry) and
/// `BookingBottomSheet` (modal, invoked from the Map's confirmed selection).
class BookingFormBody extends StatefulWidget {
  const BookingFormBody({
    super.key,
    required this.tripSelection,
    required this.apiClient,
    this.onDriverAssigned,
    this.initialCategory,
  });

  final TripSelection tripSelection;
  final ApiClient apiClient;
  final void Function(String driverId)? onDriverAssigned;

  /// Pre-selects a vehicle tier — set when the rider already tapped Bike or
  /// Car from Home before reaching this form. Null keeps the previous
  /// default (Car).
  final VehicleCategory? initialCategory;

  @override
  State<BookingFormBody> createState() => _BookingFormBodyState();
}

class _BookingFormBodyState extends State<BookingFormBody> {
  late VehicleCategory _selectedCategory;
  PaymentMethod _selectedPayment = MockBookingCatalog.paymentMethods.first;
  Voucher? _selectedVoucher;
  PromoResult? _promoResult;
  String? _voucherError;
  bool _applyingVoucher = false;
  String? _bookingError;

  FareEstimate? _fareEstimate;
  bool _loadingFare = true;

  @override
  void initState() {
    super.initState();
    _selectedCategory = widget.initialCategory ?? VehicleCategory.car;
    _loadFare();
  }

  VehicleOption get _selectedVehicle =>
      MockBookingCatalog.vehicles.firstWhere((v) => v.category == _selectedCategory);

  Future<void> _loadFare() async {
    setState(() => _loadingFare = true);
    try {
      final estimate = await PricingRepository(widget.apiClient).estimateFare(
        pickup: widget.tripSelection.pickup,
        destination: widget.tripSelection.destination,
        serviceType: _selectedCategory.backendKey,
        tripType: 'ride',
      );
      if (!mounted) return;
      setState(() {
        _fareEstimate = estimate;
        _loadingFare = false;
      });
      // The base fare just changed (vehicle/route) — any discount computed
      // against the old total is stale, so re-check the currently selected
      // voucher (if any) against the new total rather than showing a wrong number.
      if (_selectedVoucher != null) _applyVoucher(_selectedVoucher);
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _fareEstimate = null;
        _loadingFare = false;
      });
    }
  }

  Future<void> _pickVoucher() async {
    final result = await VoucherListSheet.show(context, apiClient: widget.apiClient, selected: _selectedVoucher);
    if (!mounted) return;
    await _applyVoucher(result);
  }

  /// Calls `POST /api/v1/promo/apply` against the current fare estimate's
  /// total — the backend is the only source of truth for the discount
  /// amount, matching "Pricing chỉ đọc discount từ Promotion. Không tự tính lại."
  Future<void> _applyVoucher(Voucher? voucher) async {
    if (voucher == null) {
      setState(() {
        _selectedVoucher = null;
        _promoResult = null;
        _voucherError = null;
      });
      return;
    }
    final fare = _fareEstimate;
    if (fare == null) return;

    setState(() {
      _selectedVoucher = voucher;
      _applyingVoucher = true;
      _voucherError = null;
    });
    try {
      final result = await PromotionRepository(widget.apiClient).apply(
        code: voucher.code,
        orderAmount: fare.total,
        serviceType: _selectedCategory.backendKey,
        tripType: 'ride',
      );
      if (!mounted) return;
      setState(() {
        _promoResult = result.applied ? result : null;
        _voucherError = result.applied ? null : (result.reason.isNotEmpty ? result.reason : 'Voucher không áp dụng được.');
        _applyingVoucher = false;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _promoResult = null;
        _voucherError = e.statusCode == 0 ? e.message : 'Voucher không áp dụng được.';
        _applyingVoucher = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _promoResult = null;
        _voucherError = 'Voucher không áp dụng được.';
        _applyingVoucher = false;
      });
    }
  }

  Future<void> _pickVehicle() async {
    final result = await VehicleListSheet.show(
      context,
      options: MockBookingCatalog.vehicles,
      selected: _selectedCategory,
      pickup: widget.tripSelection.pickup,
      destination: widget.tripSelection.destination,
      tripType: 'ride',
      apiClient: widget.apiClient,
    );
    if (!mounted || result == null) return;
    setState(() => _selectedCategory = result);
    _loadFare();
  }

  Future<void> _handleBookRide() async {
    setState(() => _bookingError = null);

    try {
      final repo = BookingRepository(widget.apiClient);
      final result = await repo.bookRide(
        widget.tripSelection,
        voucherCode: _promoResult != null ? _selectedVoucher?.code : null,
        orderAmount: _promoResult != null ? _fareEstimate?.total : null,
      );

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
    final fare = _fareEstimate;
    final promo = _promoResult;
    final finalTotal = promo?.finalOrderAmount;
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
        _VehicleSummaryRow(
          vehicle: _selectedVehicle,
          fare: fare,
          loading: _loadingFare,
          onTap: _pickVehicle,
        ),
        const SizedBox(height: AppSpacing.lg),
        AnimatedSwitcher(
          duration: const Duration(milliseconds: 250),
          child: _loadingFare
              ? const _FareLoadingSkeleton(key: ValueKey('fare-loading'))
              : fare != null
                  ? FareSummaryCard(
                      key: ValueKey('$_selectedCategory-${_selectedVoucher?.id}-$finalTotal'),
                      breakdown: fare,
                      distanceKm: fare.distanceKm,
                      durationMin: fare.durationMinutes,
                      voucher: promo != null ? _selectedVoucher : null,
                      discountAmount: promo?.discountAmount,
                      finalTotal: finalTotal,
                      // No surge data source exists anywhere in the backend
                      // today (see SurgeInfo doc comment) — stays null
                      // rather than being fabricated.
                      promotion: null,
                      surge: null,
                      cheaperThanCompetitorLabel: null,
                    )
                  : _FareErrorView(key: const ValueKey('fare-error'), onRetry: _loadFare),
        ),
        const SizedBox(height: AppSpacing.lg),
        PaymentMethodCard(
          selected: _selectedPayment,
          onChanged: (m) => setState(() => _selectedPayment = m),
        ),
        const SizedBox(height: AppSpacing.lg),
        VoucherPickerTile(
          selected: _selectedVoucher,
          onTap: _applyingVoucher ? () {} : _pickVoucher,
        ),
        if (_applyingVoucher) ...[
          const SizedBox(height: AppSpacing.xs),
          const Text('Đang áp dụng voucher…', style: TextStyle(color: AppColors.textSecondary, fontSize: 12)),
        ] else if (_voucherError != null) ...[
          const SizedBox(height: AppSpacing.xs),
          Text(_voucherError!, style: const TextStyle(color: AppColors.error, fontSize: 12)),
        ],
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
          child: fare != null
              ? BookRideButton(
                  label: 'Đặt ${_selectedVehicle.label} · ${formatMoney(finalTotal ?? fare.total, fare.currencyCode)}',
                  onConfirm: _handleBookRide,
                )
              : AppButton.primary(
                  label: 'Đặt ${_selectedVehicle.label}',
                  isLoading: _loadingFare,
                  onPressed: null,
                ),
        ),
      ],
    );
  }
}

/// Shown in place of the [FareSummaryCard] while the real quote is loading
/// — never a placeholder price, just a shape.
class _FareLoadingSkeleton extends StatelessWidget {
  const _FareLoadingSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return AppCard(
      child: SizedBox(
        height: 96,
        child: Center(
          child: SizedBox(
            width: 24,
            height: 24,
            child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.primary),
          ),
        ),
      ),
    );
  }
}

/// Shown when the fare estimate call fails — the backend is the single
/// source of truth for price, so a failure is reported honestly instead of
/// falling back to any locally-computed number.
class _FareErrorView extends StatelessWidget {
  const _FareErrorView({super.key, required this.onRetry});

  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Không thể tính giá',
            style: Theme.of(context).textTheme.titleSmall?.copyWith(color: AppColors.error),
          ),
          const SizedBox(height: AppSpacing.xs),
          Text(
            'Vui lòng kiểm tra kết nối và thử lại.',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(color: AppColors.textSecondary),
          ),
          const SizedBox(height: AppSpacing.sm),
          Align(
            alignment: Alignment.centerLeft,
            child: TextButton(onPressed: onRetry, child: const Text('Thử lại')),
          ),
        ],
      ),
    );
  }
}

/// Tappable summary row shown in the booking form for the currently
/// selected vehicle — opens [VehicleListSheet] on tap. Replaces the old
/// always-visible horizontal-scroll picker with a Be/Xanh SM-style
/// "tap to open the full list" interaction.
class _VehicleSummaryRow extends StatelessWidget {
  const _VehicleSummaryRow({
    required this.vehicle,
    required this.fare,
    required this.loading,
    required this.onTap,
  });

  final VehicleOption vehicle;
  final FareEstimate? fare;
  final bool loading;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final textTheme = Theme.of(context).textTheme;
    final priceText = loading ? 'Đang tính giá…' : (fare != null ? formatMoney(fare!.total, fare!.currencyCode) : '—');
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
            child: vehicle.imageAsset != null
                ? Padding(
                    padding: const EdgeInsets.all(4),
                    child: Image.asset(vehicle.imageAsset!, fit: BoxFit.contain),
                  )
                : Icon(vehicle.icon, color: AppColors.textPrimary, size: AppIconSize.lg),
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(vehicle.label, style: textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w700)),
                Text(priceText, style: textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary)),
              ],
            ),
          ),
          Icon(Icons.chevron_right, size: AppIconSize.md, color: AppColors.textTertiary),
        ],
      ),
    );
  }
}
