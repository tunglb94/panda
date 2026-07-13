import 'dart:async';

import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:go_router/go_router.dart';
import 'package:google_maps_flutter/google_maps_flutter.dart';
import '../../../../core/auth/auth_state.dart';
import '../../../../core/network/api_client.dart';
import '../../../../core/router/app_router.dart';
import '../../../../core/storage/token_storage.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_shadows.dart';
import '../../../../shared/utils/currency_format.dart';
import '../../../../shared/widgets/app_empty_state.dart';
import '../../../../shared/widgets/app_loading_view.dart';
import '../../../../shared/widgets/pressable_scale.dart';
import '../../../home/data/availability_repository.dart';
import '../../../kyc/data/kyc_repository.dart';
import '../../../kyc/domain/models/kyc_status.dart';
import '../../../location/services/location_upload_service.dart';
import '../../../trip/data/active_trip_repository.dart';
import '../../../trip/data/trip_offer_repository.dart';
import '../widgets/home_status_panel.dart';

// ─── Location state machine ───────────────────────────────────────────────────

enum _LocationStatus {
  loading,
  permissionDenied,
  permissionPermanentlyDenied,
  gpsDisabled,
  ready,
}

// ─── Page ─────────────────────────────────────────────────────────────────────

/// The Driver app's Home screen — the single most-viewed surface in the
/// app, so it gets the most design scrutiny of any screen.
///
/// Home mirrors the driver's live trip status (offline/online/searching/
/// incoming offer/picking up/waiting/in trip/awaiting payment/completed) by
/// *reading* the same `AvailabilityRepository`, `TripOfferRepository`, and
/// `ActiveTripRepository` that the Trips tab already owns — no new
/// endpoints, no new mutating calls. Accepting an offer and every trip
/// action (arrive/start/finish/rate) still only happen on the Trips tab;
/// Home's card is a live-updating preview with a single "Xem chi tiết" CTA
/// that navigates there. This keeps exactly one place in the app able to
/// mutate trip state, while making Home behave like a real dispatch home
/// screen instead of a bare online/offline toggle.
class MapPage extends StatefulWidget {
  const MapPage({
    super.key,
    required this.authState,
    required this.tokenStorage,
    required this.apiClient,
    required this.uploadService,
  });

  final AuthState authState;
  final TokenStorage tokenStorage;
  final ApiClient apiClient;
  final LocationUploadService uploadService;

  @override
  State<MapPage> createState() => _MapPageState();
}

class _MapPageState extends State<MapPage> with WidgetsBindingObserver {
  // — Location ——————————————————————————————————————————————————————————————————
  _LocationStatus _locationStatus = _LocationStatus.loading;
  LatLng? _position;
  GoogleMapController? _mapController;
  static const double _defaultZoom = 15.0;

  // — Availability ——————————————————————————————————————————————————————————————
  late final AvailabilityRepository _availRepo;
  bool _isOnline = false;
  bool _isLoadingStatus = true;
  bool _isToggling = false;
  String? _availError;

  // — KYC gating (Phần 7 — Online Switch chỉ mở khi đã Approved) ——————————————
  late final KYCRepository _kycRepo;
  bool _canGoOnline = true;

  // — Trip status mirror (read-only — see class doc) ——————————————————————————
  late final TripOfferRepository _offerRepo;
  late final ActiveTripRepository _activeTripRepo;
  HomePhase _homePhase = HomePhase.offline;
  String? _tripPickupAddress;
  String? _tripDropoffAddress;
  String? _tripFareLabel;
  int? _offerCountdown;
  bool _wasInActiveTrip = false;
  Timer? _statusTimer;
  Timer? _offerCountdownTimer;

  // ─── Lifecycle ──────────────────────────────────────────────────────────────

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _availRepo = AvailabilityRepository(widget.apiClient);
    _offerRepo = TripOfferRepository(apiClient: widget.apiClient);
    _activeTripRepo = ActiveTripRepository(apiClient: widget.apiClient);
    _kycRepo = KYCRepository(widget.apiClient);
    _resolveLocation();
    _fetchAvailability();
    _fetchKYCEligibility();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _statusTimer?.cancel();
    _offerCountdownTimer?.cancel();
    _mapController?.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed && _isOnline) {
      _pollTripStatus();
    }
  }

  // ─── Location ───────────────────────────────────────────────────────────────

  Future<void> _resolveLocation() async {
    if (!mounted) return;
    setState(() => _locationStatus = _LocationStatus.loading);

    final serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!mounted) return;
    if (!serviceEnabled) {
      setState(() => _locationStatus = _LocationStatus.gpsDisabled);
      return;
    }

    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    if (!mounted) return;

    if (permission == LocationPermission.denied) {
      setState(() => _locationStatus = _LocationStatus.permissionDenied);
      return;
    }
    if (permission == LocationPermission.deniedForever) {
      setState(() =>
          _locationStatus = _LocationStatus.permissionPermanentlyDenied);
      return;
    }

    try {
      final pos = await Geolocator.getCurrentPosition(
        locationSettings:
            const LocationSettings(accuracy: LocationAccuracy.high),
      ).timeout(const Duration(seconds: 10));
      if (!mounted) return;
      setState(() {
        _position = LatLng(pos.latitude, pos.longitude);
        _locationStatus = _LocationStatus.ready;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _locationStatus = _LocationStatus.gpsDisabled);
    }
  }

  void _recenter() {
    final pos = _position;
    if (pos == null) return;
    _mapController?.animateCamera(CameraUpdate.newLatLngZoom(pos, _defaultZoom));
  }

  // ─── Availability ────────────────────────────────────────────────────────────

  Future<void> _fetchAvailability() async {
    try {
      final result = await _availRepo.getAvailability();
      if (mounted) setState(() => _isOnline = result.isOnline);
      if (result.isOnline) {
        unawaited(widget.uploadService.start());
        _startStatusPolling();
      } else if (mounted) {
        setState(() => _homePhase = HomePhase.offline);
      }
    } catch (_) {
      // Non-fatal — default to offline
    } finally {
      if (mounted) setState(() => _isLoadingStatus = false);
    }
  }

  /// Phần 7/11 — best-effort: never blocks Home from loading, just gates the
  /// Switch once known. The real enforcement (including auto-expiring a
  /// document that's passed its date) is server-side in `GoOnlineUseCase`
  /// (Online Guard) — this mirrors the same "documents not expired" check
  /// client-side so the Switch shows disabled proactively, before the
  /// driver even attempts to go online, rather than only after a failed
  /// attempt flips the verification to Expired server-side.
  Future<void> _fetchKYCEligibility() async {
    try {
      final driver = await _kycRepo.getDriverVerification();
      final vehicle = await _kycRepo.getVehicleVerification();
      var eligible = (driver?.status.isApproved ?? false) &&
          (vehicle?.status.isApproved ?? false);
      if (eligible) {
        final documents = await _kycRepo.listDocuments();
        eligible = !documents.any((d) => d.uploaded && d.expired);
      }
      if (mounted) setState(() => _canGoOnline = eligible);
    } catch (_) {
      // Non-fatal — leave the switch enabled and let the server-side guard
      // be the source of truth if this check couldn't complete.
    }
  }

  Future<void> _toggle() async {
    if (_isToggling) return;
    setState(() {
      _isToggling = true;
      _availError = null;
    });
    try {
      final result = _isOnline
          ? await _availRepo.goOffline()
          : await _availRepo.goOnline();
      if (!mounted) return;
      setState(() => _isOnline = result.isOnline);
      if (result.isOnline) {
        unawaited(widget.uploadService.start());
        _startStatusPolling();
      } else {
        widget.uploadService.stop();
        _stopStatusPolling();
        setState(() => _homePhase = HomePhase.offline);
      }
    } on ApiException catch (e) {
      if (mounted) setState(() => _availError = e.message);
    } catch (_) {
      if (mounted) {
        setState(
            () => _availError = 'Không thể cập nhật trạng thái. Vui lòng thử lại.');
      }
    } finally {
      if (mounted) setState(() => _isToggling = false);
    }
  }

  // ─── Trip status mirror (read-only) ──────────────────────────────────────────

  void _startStatusPolling() {
    _statusTimer?.cancel();
    _pollTripStatus();
    // Lighter cadence than the Trips tab's 5s poll (this is a display
    // mirror, not the flow that drives trip-critical timing).
    _statusTimer = Timer.periodic(const Duration(seconds: 8), (_) => _pollTripStatus());
  }

  void _stopStatusPolling() {
    _statusTimer?.cancel();
    _statusTimer = null;
    _offerCountdownTimer?.cancel();
    _offerCountdownTimer = null;
  }

  Future<void> _pollTripStatus() async {
    if (!_isOnline || !mounted) return;
    try {
      final storedId = await _activeTripRepo.getStoredTripId();
      if (storedId != null) {
        final trip = await _activeTripRepo.fetchTrip(storedId);
        if (!mounted) return;
        _wasInActiveTrip = true;
        _offerCountdownTimer?.cancel();
        setState(() {
          _tripPickupAddress = trip.pickupAddress;
          _tripDropoffAddress = trip.dropoffAddress;
          _tripFareLabel = _fareLabel(trip.finalFare, trip.fareCurrency);
          _homePhase = switch (trip.status) {
            'driver_assigned' => HomePhase.pickingUp,
            'driver_arrived' => HomePhase.waiting,
            'in_progress' => HomePhase.inTrip,
            'payment_pending' || 'payment_success' => HomePhase.awaitingPayment,
            _ => HomePhase.online,
          };
        });
        return;
      }

      // No active trip in storage right now.
      if (_wasInActiveTrip) {
        // We had one last poll and don't anymore — the Trips tab just
        // settled it. Flash "Completed" briefly before returning to the
        // normal online/searching state (see class doc for why this is a
        // short client-side heuristic rather than a real observed state).
        _wasInActiveTrip = false;
        if (!mounted) return;
        setState(() => _homePhase = HomePhase.completed);
        Future.delayed(const Duration(seconds: 3), () {
          if (mounted && _homePhase == HomePhase.completed) {
            setState(() => _homePhase = HomePhase.online);
          }
        });
        return;
      }

      final offer = await _offerRepo.getCurrentOffer();
      if (!mounted) return;
      if (offer != null) {
        setState(() {
          _tripPickupAddress = offer.pickupAddress;
          _tripDropoffAddress = offer.dropoffAddress;
          _homePhase = HomePhase.incomingTrip;
        });
        _startOfferCountdown(offer.offerExpiresAt);
      } else {
        _offerCountdownTimer?.cancel();
        setState(() => _homePhase = HomePhase.online);
      }
    } catch (_) {
      // Non-fatal — keep the last known phase, retry next tick.
    }
  }

  void _startOfferCountdown(DateTime expiresAt) {
    _offerCountdownTimer?.cancel();
    void tick() {
      if (!mounted) return;
      final remaining = expiresAt.difference(DateTime.now().toUtc()).inSeconds;
      setState(() => _offerCountdown = remaining.clamp(0, 999));
    }
    tick();
    _offerCountdownTimer = Timer.periodic(const Duration(seconds: 1), (_) => tick());
  }

  static String _fareLabel(int finalFare, String currency) {
    if (finalFare > 0 && currency.isNotEmpty) {
      return formatMoney(finalFare, currency);
    }
    return '—';
  }

  void _goToTrips() => context.go(AppRoutes.trips);
  void _goToProfile() => context.go(AppRoutes.profile);

  // ─── Build ────────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: switch (_locationStatus) {
        _LocationStatus.loading =>
          const AppLoadingView(label: 'Đang xác định vị trí của bạn…'),
        _LocationStatus.permissionDenied => _LocationErrorView(
            icon: Icons.location_off,
            title: 'Quyền truy cập vị trí bị từ chối',
            message:
                'PandaDriver cần quyền truy cập vị trí để hiển thị vị trí '
                'của bạn trên bản đồ.',
            actionLabel: 'Cấp quyền',
            onAction: _resolveLocation,
          ),
        _LocationStatus.permissionPermanentlyDenied => _LocationErrorView(
            icon: Icons.location_disabled,
            title: 'Vị trí đang bị chặn',
            message:
                'Vui lòng bật quyền truy cập vị trí cho PandaDriver trong '
                'phần Cài đặt của thiết bị.',
            actionLabel: 'Mở Cài đặt',
            onAction: () async => Geolocator.openAppSettings(),
          ),
        _LocationStatus.gpsDisabled => _LocationErrorView(
            icon: Icons.gps_off,
            title: 'GPS đang tắt',
            message:
                'Hãy bật Dịch vụ vị trí để PandaDriver có thể hiển thị vị '
                'trí của bạn trên bản đồ.',
            actionLabel: 'Mở Cài đặt vị trí',
            onAction: () async => Geolocator.openLocationSettings(),
          ),
        _LocationStatus.ready => _buildMap(),
      },
    );
  }

  Widget _buildMap() {
    return Stack(
      children: [
        GoogleMap(
          initialCameraPosition: CameraPosition(
            target: _position!,
            zoom: _defaultZoom,
          ),
          onMapCreated: (controller) => _mapController = controller,
          myLocationEnabled: true,
          myLocationButtonEnabled: false,
          zoomControlsEnabled: false,
          compassEnabled: true,
          mapToolbarEnabled: false,
          mapType: MapType.normal,
          // Reserve space above the status panel so map controls/markers
          // near the bottom stay visible and tappable.
          padding: const EdgeInsets.only(bottom: 200),
        ),
        SafeArea(
          bottom: false,
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                _DriverAvatarButton(onTap: _goToProfile),
                _RecenterButton(onTap: _recenter),
              ],
            ),
          ),
        ),
        Positioned(
          left: 0,
          right: 0,
          bottom: 0,
          child: HomeStatusPanel(
            phase: _homePhase,
            isBusy: _isLoadingStatus || _isToggling,
            error: _availError,
            pickupAddress: _tripPickupAddress,
            dropoffAddress: _tripDropoffAddress,
            countdownSeconds: _offerCountdown,
            fareLabel: _tripFareLabel,
            canGoOnline: _canGoOnline,
            onToggleOnline: _toggle,
            onViewTrip: _goToTrips,
          ),
        ),
      ],
    );
  }
}

// ─── Top bar affordances ───────────────────────────────────────────────────────

class _DriverAvatarButton extends StatelessWidget {
  const _DriverAvatarButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return _CircleFloatingButton(
      onTap: onTap,
      icon: Icons.person,
      iconSize: 24,
    );
  }
}

class _RecenterButton extends StatelessWidget {
  const _RecenterButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return _CircleFloatingButton(
      onTap: onTap,
      icon: Icons.my_location,
      iconSize: 22,
    );
  }
}

/// Shared shell for the small circular floating buttons on Home (avatar,
/// recenter): press-down scale (via `InkWell.onHighlightChanged`, which is
/// always safe to combine with `InkWell`'s own tap handling — no nested
/// gesture detectors), soft card shadow, ripple.
class _CircleFloatingButton extends StatefulWidget {
  const _CircleFloatingButton({
    required this.onTap,
    required this.icon,
    required this.iconSize,
  });

  final VoidCallback onTap;
  final IconData icon;
  final double iconSize;

  @override
  State<_CircleFloatingButton> createState() => _CircleFloatingButtonState();
}

class _CircleFloatingButtonState extends State<_CircleFloatingButton> {
  bool _pressed = false;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: AppColors.surface,
      shape: const CircleBorder(),
      elevation: 0,
      child: InkWell(
        onTap: widget.onTap,
        customBorder: const CircleBorder(),
        onHighlightChanged: (v) => setState(() => _pressed = v),
        child: PressableScale(
          pressed: _pressed,
          child: Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              boxShadow: AppShadows.card,
            ),
            alignment: Alignment.center,
            child: Icon(widget.icon, color: AppColors.primary, size: widget.iconSize),
          ),
        ),
      ),
    );
  }
}

// ─── Error / blocked view ─────────────────────────────────────────────────────

class _LocationErrorView extends StatelessWidget {
  const _LocationErrorView({
    required this.icon,
    required this.title,
    required this.message,
    required this.actionLabel,
    required this.onAction,
  });

  final IconData icon;
  final String title;
  final String message;
  final String actionLabel;
  final VoidCallback onAction;

  @override
  Widget build(BuildContext context) {
    return AppEmptyState(
      icon: icon,
      title: title,
      subtitle: message,
      actionLabel: actionLabel,
      onAction: onAction,
    );
  }
}
