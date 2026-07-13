import 'package:flutter/material.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/places/nominatim_places_service.dart';
import 'package:rider/core/theme/app_colors.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/booking/data/pricing_repository.dart';
import 'package:rider/features/booking/domain/models/fare_estimate.dart';
import 'package:rider/features/booking/domain/models/mock_booking_catalog.dart';
import 'package:rider/features/booking/domain/models/vehicle_option.dart';
import 'package:rider/features/delivery/data/delivery_repository.dart';
import 'package:rider/features/delivery/domain/models/package_category.dart';
import 'package:rider/features/delivery/presentation/pages/delivery_lifecycle_page.dart';
import 'package:rider/features/map/presentation/pages/map_address_picker_page.dart';
import 'package:rider/features/map/presentation/widgets/place_search_field.dart';
import 'package:rider/shared/utils/currency_format.dart';
import 'package:rider/shared/widgets/app_button.dart';
import 'package:rider/shared/widgets/app_card.dart';
import 'package:rider/shared/widgets/app_snackbar.dart';

/// "Gửi hàng" — Phase 1 of the Delivery production pass. Sender fills in
/// pickup/receiver addresses, receiver contact, note, item type, and
/// declared value, sees a live estimate, then books. Mirrors
/// `BookingFormBody`'s architecture (same shared widgets, same
/// estimate-then-book shape) but is an entirely separate page/flow — Ride's
/// booking screen is untouched.
class DeliveryFormPage extends StatefulWidget {
  const DeliveryFormPage({super.key, required this.apiClient, this.initialBias});

  final ApiClient apiClient;

  /// The rider's last known position, when the caller already has one
  /// (`MapPage` always does by the time it can push this page) — used to
  /// bias address-suggestion ranking toward nearby results, same as Ride's
  /// own search fields already do. Null when opened from a context with no
  /// known position yet (e.g. the Home Hub) — suggestions are then
  /// unbiased, exactly as this page behaved before this fix.
  final LatLng? initialBias;

  @override
  State<DeliveryFormPage> createState() => _DeliveryFormPageState();
}

class _DeliveryFormPageState extends State<DeliveryFormPage> {
  static const _placesService = NominatimPlacesService();

  String? _pickupAddress;
  LatLng? _pickupLocation;
  String? _receiverAddress;
  LatLng? _receiverLocation;

  final _senderNameController = TextEditingController();
  final _receiverNameController = TextEditingController();
  final _receiverPhoneController = TextEditingController();
  final _noteController = TextEditingController();
  final _valueController = TextEditingController();

  PackageCategory _category = PackageCategory.small;
  VehicleOption _vehicle = MockBookingCatalog.vehicles.first;
  bool _booking = false;

  FareEstimate? _fareEstimate;
  bool _loadingFare = false;

  @override
  void dispose() {
    _senderNameController.dispose();
    _receiverNameController.dispose();
    _receiverPhoneController.dispose();
    _noteController.dispose();
    _valueController.dispose();
    super.dispose();
  }

  Future<void> _pickFromMap({required bool isPickup}) async {
    final result = await Navigator.of(context).push<(String, LatLng)>(
      MaterialPageRoute(
        builder: (_) => MapAddressPickerPage(
          initialPosition: widget.initialBias,
          title: isPickup ? 'Chọn điểm lấy hàng' : 'Chọn điểm giao hàng',
        ),
      ),
    );
    if (result == null || !mounted) return;
    final (address, location) = result;
    setState(() {
      if (isPickup) {
        _pickupAddress = address;
        _pickupLocation = location;
      } else {
        _receiverAddress = address;
        _receiverLocation = location;
      }
    });
    _loadFare();
  }

  bool get _hasRoute => _pickupLocation != null && _receiverLocation != null;

  bool get _canBook =>
      _hasRoute &&
      _fareEstimate != null &&
      _receiverNameController.text.trim().isNotEmpty &&
      _receiverPhoneController.text.trim().isNotEmpty;

  /// Fetches the real fare from the backend's Pricing service
  /// (`PricingRepository.estimateFare`) — no client-side fare math. A failed
  /// call leaves [_fareEstimate] null; [_canBook] then stays false, so
  /// booking is blocked rather than proceeding on a guessed price.
  Future<void> _loadFare() async {
    if (!_hasRoute) {
      setState(() => _fareEstimate = null);
      return;
    }
    setState(() => _loadingFare = true);
    try {
      final estimate = await PricingRepository(widget.apiClient).estimateFare(
        pickup: _pickupLocation!,
        destination: _receiverLocation!,
        serviceType: _vehicle.category.backendKey,
        tripType: 'delivery',
      );
      if (!mounted) return;
      setState(() {
        _fareEstimate = estimate;
        _loadingFare = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _fareEstimate = null;
        _loadingFare = false;
      });
    }
  }

  /// The wire's only free-text channel for what's in the package — no
  /// `package_type` field exists on `CreateTripRequest`/`BookRideRequest`
  /// (see the Delivery wire-contract audit), so the chosen category is
  /// prefixed onto the real `package_note` the driver will actually read,
  /// never sent as a fabricated structured field.
  String get _composedNote {
    final note = _noteController.text.trim();
    final prefix = '[${_category.label}]';
    return note.isEmpty ? prefix : '$prefix $note';
  }

  Future<void> _book() async {
    if (!_canBook || _booking) return;
    final fare = _fareEstimate;
    if (fare == null) return;
    setState(() => _booking = true);
    try {
      final valueVnd = int.tryParse(_valueController.text.replaceAll('.', '').trim()) ?? 0;
      final result = await DeliveryRepository(widget.apiClient).bookDelivery(
        pickupAddress: _pickupAddress!,
        pickupLat: _pickupLocation!.latitude,
        pickupLon: _pickupLocation!.longitude,
        receiverAddress: _receiverAddress!,
        pickupContactName: _senderNameController.text.trim(),
        receiverName: _receiverNameController.text.trim(),
        receiverPhone: _receiverPhoneController.text.trim(),
        packageNote: _composedNote,
        packageValueCents: valueVnd,
      );
      if (!mounted) return;
      Navigator.of(context).pushReplacement(MaterialPageRoute(
        builder: (_) => DeliveryLifecyclePage(
          tripId: result.tripId,
          apiClient: widget.apiClient,
          pickupAddress: _pickupAddress!,
          pickupLocation: _pickupLocation!,
          receiverAddress: _receiverAddress!,
          receiverLocation: _receiverLocation!,
          receiverName: _receiverNameController.text.trim(),
          receiverPhone: _receiverPhoneController.text.trim(),
          estimatedFareCents: fare.total,
          currency: fare.currencyCode,
        ),
      ));
    } on ApiException catch (e) {
      if (mounted) {
        AppSnackbar.error(context, e.statusCode == 0 ? e.message : 'Đặt đơn giao hàng thất bại. Vui lòng thử lại.');
      }
    } catch (_) {
      if (mounted) AppSnackbar.error(context, 'Đặt đơn giao hàng thất bại. Vui lòng thử lại.');
    } finally {
      if (mounted) setState(() => _booking = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      appBar: AppBar(title: const Text('Gửi hàng')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: ListView(
              padding: const EdgeInsets.all(AppSpacing.lg),
              children: [
                Text('Điểm lấy & giao hàng', style: theme.textTheme.titleSmall),
                const SizedBox(height: AppSpacing.sm),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: PlaceSearchField(
                        key: ValueKey('pickup-$_pickupAddress'),
                        placesService: _placesService,
                        hintText: 'Điểm lấy hàng',
                        biasCenter: widget.initialBias,
                        initialText: _pickupAddress,
                        onSelected: (address, location) => setState(() {
                          _pickupAddress = address;
                          _pickupLocation = location;
                        }),
                      ),
                    ),
                    const SizedBox(width: AppSpacing.xs),
                    IconButton(
                      tooltip: 'Chọn điểm lấy hàng từ bản đồ',
                      icon: const Icon(Icons.map_outlined),
                      onPressed: () => _pickFromMap(isPickup: true),
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.sm),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: PlaceSearchField(
                        key: ValueKey('dropoff-$_receiverAddress'),
                        placesService: _placesService,
                        hintText: 'Điểm giao hàng',
                        biasCenter: widget.initialBias,
                        initialText: _receiverAddress,
                        onSelected: (address, location) => setState(() {
                          _receiverAddress = address;
                          _receiverLocation = location;
                        }),
                      ),
                    ),
                    const SizedBox(width: AppSpacing.xs),
                    IconButton(
                      tooltip: 'Chọn điểm giao hàng từ bản đồ',
                      icon: const Icon(Icons.map_outlined),
                      onPressed: () => _pickFromMap(isPickup: false),
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.xl),
                Text('Người gửi', style: theme.textTheme.titleSmall),
                const SizedBox(height: AppSpacing.sm),
                TextField(
                  controller: _senderNameController,
                  decoration: const InputDecoration(
                    hintText: 'Tên người gửi',
                    border: OutlineInputBorder(),
                    isDense: true,
                  ),
                ),
                const SizedBox(height: AppSpacing.xl),
                Text('Người nhận', style: theme.textTheme.titleSmall),
                const SizedBox(height: AppSpacing.sm),
                TextField(
                  controller: _receiverNameController,
                  onChanged: (_) => setState(() {}),
                  decoration: const InputDecoration(
                    hintText: 'Tên người nhận',
                    border: OutlineInputBorder(),
                    isDense: true,
                  ),
                ),
                const SizedBox(height: AppSpacing.sm),
                TextField(
                  controller: _receiverPhoneController,
                  onChanged: (_) => setState(() {}),
                  keyboardType: TextInputType.phone,
                  decoration: const InputDecoration(
                    hintText: 'Số điện thoại người nhận',
                    border: OutlineInputBorder(),
                    isDense: true,
                  ),
                ),
                const SizedBox(height: AppSpacing.xl),
                Text('Loại hàng', style: theme.textTheme.titleSmall),
                const SizedBox(height: AppSpacing.sm),
                Wrap(
                  spacing: AppSpacing.sm,
                  runSpacing: AppSpacing.sm,
                  children: [
                    for (final category in PackageCategory.values)
                      ChoiceChip(
                        label: Text(category.label),
                        avatar: Icon(category.icon, size: 18),
                        selected: _category == category,
                        onSelected: (_) => setState(() => _category = category),
                      ),
                  ],
                ),
                const SizedBox(height: AppSpacing.xl),
                Text('Ghi chú', style: theme.textTheme.titleSmall),
                const SizedBox(height: AppSpacing.sm),
                TextField(
                  controller: _noteController,
                  maxLines: 2,
                  maxLength: 200,
                  decoration: const InputDecoration(
                    hintText: 'Ghi chú cho tài xế (không bắt buộc)',
                    border: OutlineInputBorder(),
                  ),
                ),
                Text('Giá trị hàng (đ)', style: theme.textTheme.titleSmall),
                const SizedBox(height: AppSpacing.sm),
                TextField(
                  controller: _valueController,
                  keyboardType: TextInputType.number,
                  decoration: const InputDecoration(
                    hintText: 'Giá trị hàng — không bắt buộc',
                    border: OutlineInputBorder(),
                    isDense: true,
                  ),
                ),
                const SizedBox(height: AppSpacing.xl),
                Text('Loại xe', style: theme.textTheme.titleSmall),
                const SizedBox(height: AppSpacing.sm),
                Wrap(
                  spacing: AppSpacing.sm,
                  runSpacing: AppSpacing.sm,
                  children: [
                    for (final v in MockBookingCatalog.vehicles.where((v) => v.isAvailable))
                      ChoiceChip(
                        label: Text(v.label),
                        avatar: Icon(v.icon, size: 18),
                        selected: _vehicle.category == v.category,
                        onSelected: (_) {
                          setState(() => _vehicle = v);
                          _loadFare();
                        },
                      ),
                  ],
                ),
                const SizedBox(height: AppSpacing.xl),
                if (!_hasRoute)
                  Text(
                    'Chọn điểm lấy và giao hàng để xem giá ước tính.',
                    style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textTertiary),
                  )
                else if (_loadingFare)
                  const AppCard(
                    child: SizedBox(
                      height: 64,
                      child: Center(
                        child: SizedBox(
                          width: 24,
                          height: 24,
                          child: CircularProgressIndicator(strokeWidth: 2, color: AppColors.primary),
                        ),
                      ),
                    ),
                  )
                else if (_fareEstimate == null)
                  AppCard(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Không thể tính giá',
                          style: theme.textTheme.titleSmall?.copyWith(color: AppColors.error),
                        ),
                        const SizedBox(height: AppSpacing.xs),
                        Align(
                          alignment: Alignment.centerLeft,
                          child: TextButton(onPressed: _loadFare, child: const Text('Thử lại')),
                        ),
                      ],
                    ),
                  )
                else
                  AppCard(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text('Giá ước tính', style: theme.textTheme.bodyMedium
                                ?.copyWith(color: AppColors.textSecondary)),
                            Flexible(
                              child: Text(
                                formatMoney(_fareEstimate!.total, _fareEstimate!.currencyCode),
                                textAlign: TextAlign.right,
                                style: theme.textTheme.titleLarge?.copyWith(color: AppColors.primary),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 4),
                        Text(
                          '${_fareEstimate!.distanceKm.toStringAsFixed(1)} km · ${_fareEstimate!.durationMinutes.round()} phút',
                          style: theme.textTheme.labelMedium?.copyWith(color: AppColors.textTertiary),
                        ),
                      ],
                    ),
                  ),
                const SizedBox(height: AppSpacing.lg),
                AppButton.primary(
                  label: 'Đặt đơn',
                  icon: Icons.local_shipping_outlined,
                  isLoading: _booking,
                  onPressed: _canBook ? _book : null,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
