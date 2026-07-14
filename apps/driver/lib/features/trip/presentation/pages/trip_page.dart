import 'dart:async';

import 'package:audioplayers/audioplayers.dart';
import 'package:flutter/material.dart';
import 'package:vibration/vibration.dart';

import '../../../../core/location/location_engine.dart';
import '../../../../core/network/api_client.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_icon_sizes.dart';
import '../../../../core/theme/app_radius.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../core/trip_metrics/trip_metrics.dart';
import '../../../../core/trip_metrics/trip_metrics_engine.dart';
import '../../../../shared/utils/currency_format.dart';
import '../../../../shared/widgets/app_button.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_dialog.dart';
import '../../../../shared/widgets/app_empty_state.dart';
import '../../../../shared/widgets/app_loading_view.dart';
import '../../../../shared/widgets/app_status_chip.dart';
import '../../../../shared/widgets/fare_breakdown_waterfall.dart';
import '../../../../shared/widgets/mascot_image.dart';
import '../../data/active_trip_repository.dart';
import '../../data/trip_offer_repository.dart';
import '../widgets/delivery_execution_card.dart';
import '../widgets/delivery_offer_card.dart';
import '../widgets/passenger_info_card.dart';
import '../widgets/sos_button.dart';
import '../widgets/trip_timeline.dart';

enum _PageState {
  initializing,
  polling,
  offerAvailable,
  acting,
  activeTrip,
  awaitingPayment,
  completed,
  error,
}

class TripPage extends StatefulWidget {
  const TripPage({
    super.key,
    required this.apiClient,
    required this.locationStream,
  });

  final ApiClient apiClient;
  final Stream<LocationUpdate> locationStream;

  @override
  State<TripPage> createState() => _TripPageState();
}

class _TripPageState extends State<TripPage> with WidgetsBindingObserver {
  late final TripOfferRepository _offerRepo;
  late final ActiveTripRepository _activeTripRepo;
  late final TripMetricsEngine _metricsEngine;

  _PageState _state = _PageState.initializing;
  TripOffer? _offer;
  ActiveTrip? _activeTrip;
  String? _errorMessage;
  String _actingLabel = 'Vui lòng chờ…';
  int _countdownSeconds = 0;
  bool _hasArrived = false;
  TripMetrics? _finalMetrics;
  String? _completedTripId;

  /// True for a brief moment right after an offer accept succeeds — swaps
  /// the [_PageState.acting] screen from a spinner to a checkmark flash
  /// before continuing to [_PageState.activeTrip]. Purely a presentation
  /// detail layered onto the existing state machine; the transition target
  /// and every API call are unchanged.
  bool _showAcceptSuccess = false;

  Timer? _pollTimer;
  Timer? _countdownTimer;
  Timer? _paymentPollTimer;
  bool _isPollingActive = false;
  bool _isPaymentPollingActive = false;
  final _offerSoundPlayer = AudioPlayer();

  /// Plays the pickup alert sound + vibrates so the driver notices a new
  /// trip offer even if the phone isn't in hand.
  Future<void> _notifyNewOffer() async {
    unawaited(_offerSoundPlayer.play(AssetSource('sounds/panda_pickup.mp3')));
    if (await Vibration.hasVibrator()) {
      unawaited(Vibration.vibrate(pattern: [0, 400, 200, 400]));
    }
  }

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    _offerRepo = TripOfferRepository(apiClient: widget.apiClient);
    _activeTripRepo = ActiveTripRepository(apiClient: widget.apiClient);
    _metricsEngine = TripMetricsEngine(locationStream: widget.locationStream);
    _initialize();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    _pollTimer?.cancel();
    _countdownTimer?.cancel();
    _paymentPollTimer?.cancel();
    _metricsEngine.reset();
    _offerSoundPlayer.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state != AppLifecycleState.resumed) return;
    switch (_state) {
      case _PageState.polling:
        _poll();
      case _PageState.activeTrip || _PageState.awaitingPayment:
        _resumeActiveTrip();
      default:
        break;
    }
  }

  Future<void> _resumeActiveTrip() async {
    final tripId = _activeTrip?.tripId;
    if (tripId == null) return;
    try {
      final trip = await _activeTripRepo.fetchTrip(tripId);
      if (!mounted) return;
      setState(() => _activeTrip = trip);
      if (trip.isAwaitingPayment && _state == _PageState.activeTrip) {
        _paymentPollTimer?.cancel();
        setState(() => _state = _PageState.awaitingPayment);
        _startPaymentPolling();
      }
    } on ApiException catch (_) {
      // Ignore; regular timers will handle retry
    }
  }

  // ─── Initialization ──────────────────────────────────────────────────────────

  Future<void> _initialize() async {
    final storedId = await _activeTripRepo.getStoredTripId();
    if (!mounted) return;

    if (storedId != null) {
      try {
        final trip = await _activeTripRepo.fetchTrip(storedId);
        if (!mounted) return;
        if (trip.isActive) {
          setState(() {
            _state = _PageState.activeTrip;
            _activeTrip = trip;
            _hasArrived = trip.status == 'driver_arrived';
          });
          return;
        }
        if (trip.isAwaitingPayment) {
          setState(() {
            _state = _PageState.awaitingPayment;
            _activeTrip = trip;
          });
          _startPaymentPolling();
          return;
        }
        // Trip completed, settled, or cancelled on backend — clear and fall through.
        await _activeTripRepo.clearActiveTripId();
      } on ApiException catch (e) {
        if (!mounted) return;
        if (e.statusCode == 404) {
          await _activeTripRepo.clearActiveTripId();
          // Fall through to polling.
        } else {
          setState(() {
            _state = _PageState.error;
            _errorMessage = _friendlyError(e);
          });
          return;
        }
      }
    }

    if (!mounted) return;
    _startPolling();
  }

  // ─── Offer polling ───────────────────────────────────────────────────────────

  void _startPolling() {
    _pollTimer?.cancel();
    setState(() {
      _state = _PageState.polling;
      _offer = null;
    });
    _poll();
    _pollTimer = Timer.periodic(const Duration(seconds: 5), (_) => _poll());
  }

  Future<void> _poll() async {
    if (_isPollingActive) return;
    if (_state == _PageState.acting ||
        _state == _PageState.activeTrip ||
        _state == _PageState.awaitingPayment ||
        _state == _PageState.completed ||
        _state == _PageState.initializing) {
      return;
    }

    _isPollingActive = true;
    try {
      final offer = await _offerRepo.getCurrentOffer();
      if (!mounted) return;
      if (offer == null) {
        _countdownTimer?.cancel();
        if (_state != _PageState.polling) {
          setState(() {
            _state = _PageState.polling;
            _offer = null;
          });
        }
      } else {
        if (_state != _PageState.offerAvailable ||
            _offer?.tripId != offer.tripId) {
          _startCountdown(offer);
          setState(() {
            _state = _PageState.offerAvailable;
            _offer = offer;
          });
          _notifyNewOffer();
        }
      }
    } on ApiException catch (e) {
      if (!mounted) return;
      if (_state != _PageState.error) {
        setState(() {
          _state = _PageState.error;
          _errorMessage = _friendlyError(e);
        });
      }
    } finally {
      _isPollingActive = false;
    }
  }

  // ─── Payment polling ─────────────────────────────────────────────────────────

  void _startPaymentPolling() {
    _paymentPollTimer?.cancel();
    _paymentPollTimer =
        Timer.periodic(const Duration(seconds: 3), (_) => _paymentPoll());
  }

  Future<void> _paymentPoll() async {
    // Overlap guard — mirrors _poll()'s _isPollingActive. Without this, a
    // response slower than the 3s tick interval could leave two fetchTrip
    // calls in flight; a stale second response landing after the first
    // already drove the page to `completed` would otherwise overwrite
    // _activeTrip with outdated data or re-run the settle transition.
    if (_isPaymentPollingActive) return;
    final trip = _activeTrip;
    if (trip == null) return;
    _isPaymentPollingActive = true;
    try {
      final updated = await _activeTripRepo.fetchTrip(trip.tripId);
      if (!mounted) return;
      if (_state != _PageState.awaitingPayment) return;
      if (updated.status == 'settled') {
        _paymentPollTimer?.cancel();
        _paymentPollTimer = null;
        await _activeTripRepo.clearActiveTripId();
        if (!mounted) return;
        setState(() {
          _state = _PageState.completed;
          _activeTrip = updated;
          _completedTripId = updated.tripId;
        });
      } else {
        setState(() => _activeTrip = updated);
      }
    } on ApiException catch (_) {
      // Ignore transient poll errors; rider may still be paying.
    } finally {
      _isPaymentPollingActive = false;
    }
  }

  /// Never show a raw backend error string to the driver. `statusCode == 0`
  /// is only ever thrown client-side by [ApiClient] itself (timeout/
  /// connectivity) with copy that's already Vietnamese and safe to show
  /// verbatim; any real HTTP status carries a raw backend message instead.
  String _friendlyError(ApiException e) =>
      e.statusCode == 0 ? e.message : 'Không thể hoàn tất thao tác. Vui lòng thử lại.';

  void _onRatingDone() {
    _startPolling();
  }

  Future<void> _submitDriverRating(int stars, String? comment) async {
    final tripId = _completedTripId;
    final riderId = _activeTrip?.riderId;
    if (tripId != null && riderId != null && riderId.isNotEmpty) {
      try {
        final body = <String, dynamic>{
          'stars': stars,
          'ratee_id': riderId,
          'role': 'driver',
        };
        if (comment != null && comment.isNotEmpty) body['comment'] = comment;
        await widget.apiClient.post('/api/v1/rides/$tripId/rate', body: body);
      } catch (_) {
        // Non-fatal: proceed to offer queue regardless.
      }
    }
    if (mounted) _onRatingDone();
  }

  void _startCountdown(TripOffer offer) {
    _countdownTimer?.cancel();
    _countdownSeconds =
        offer.offerExpiresAt.difference(DateTime.now().toUtc()).inSeconds;
    if (_countdownSeconds < 0) _countdownSeconds = 0;
    _countdownTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (!mounted) return;
      setState(() {
        _countdownSeconds =
            (_offer?.offerExpiresAt.difference(DateTime.now().toUtc()).inSeconds ??
                    0)
                .clamp(0, 999);
      });
      if (_countdownSeconds == 0) {
        _countdownTimer?.cancel();
        if (mounted && _state == _PageState.offerAvailable) {
          setState(() {
            _state = _PageState.polling;
            _offer = null;
          });
        }
      }
    });
  }

  // ─── Offer actions ───────────────────────────────────────────────────────────

  Future<void> _onAccept() async {
    // Synchronous re-entry guard — checked before any `await`, so a rapid
    // double-tap (both onPressed calls landing before the first rebuild
    // disables the button) cannot fire two AcceptDispatchOffer requests.
    if (_state == _PageState.acting) return;
    final offer = _offer;
    if (offer == null) return;
    _pollTimer?.cancel();
    _countdownTimer?.cancel();
    setState(() { _actingLabel = 'Đang chấp nhận chuyến…'; _state = _PageState.acting; });
    try {
      await _offerRepo.acceptOffer(offer.tripId);
      await _activeTripRepo.saveActiveTripId(offer.tripId);
      if (!mounted) return;
      setState(() => _showAcceptSuccess = true);
      await Future.delayed(const Duration(milliseconds: 550));
      if (!mounted) return;
      setState(() {
        _showAcceptSuccess = false;
        _state = _PageState.activeTrip;
        _activeTrip = ActiveTrip(
          tripId: offer.tripId,
          pickupAddress: offer.pickupAddress,
          dropoffAddress: offer.dropoffAddress,
          status: 'driver_assigned',
          tripType: offer.tripType,
        );
        _hasArrived = false;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = _friendlyError(e);
      });
    }
  }

  Future<void> _onReject() async {
    if (_state == _PageState.acting) return;
    final offer = _offer;
    if (offer == null) return;
    _pollTimer?.cancel();
    _countdownTimer?.cancel();
    setState(() { _actingLabel = 'Đang từ chối…'; _state = _PageState.acting; });
    try {
      await _offerRepo.rejectOffer(offer.tripId);
      if (!mounted) return;
      _startPolling();
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = _friendlyError(e);
      });
    }
  }

  // ─── Trip execution actions ──────────────────────────────────────────────────

  Future<void> _onArrived() async {
    if (_state == _PageState.acting) return;
    final trip = _activeTrip;
    if (trip == null) return;
    setState(() { _actingLabel = 'Đang xác nhận đã đến…'; _state = _PageState.acting; });
    try {
      await _activeTripRepo.arriveAtPickup(trip.tripId);
      if (!mounted) return;
      setState(() {
        _state = _PageState.activeTrip;
        _activeTrip = trip.copyWith(status: 'driver_arrived');
        _hasArrived = true;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = _friendlyError(e);
      });
    }
  }

  Future<void> _onStartTrip() async {
    if (_state == _PageState.acting) return;
    final trip = _activeTrip;
    if (trip == null) return;
    setState(() { _actingLabel = 'Đang bắt đầu chuyến đi…'; _state = _PageState.acting; });
    try {
      await _activeTripRepo.startTrip(trip.tripId);
      if (!mounted) return;
      _metricsEngine.start();
      setState(() {
        _state = _PageState.activeTrip;
        _activeTrip = trip.copyWith(status: 'in_progress');
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = _friendlyError(e);
      });
    }
  }

  // ─── Delivery lifecycle actions ─────────────────────────────────────────────
  // Delivery V1: Accept (existing _onAccept, shared with Ride) → Arrive
  // Pickup (existing _onArrived, shared — Trip's arrive/MarkDriverArrived is
  // reused unchanged for Delivery) → Pickup Parcel → Start Delivery →
  // Complete Delivery. The last 3 are Delivery-only, calling the Trip
  // service directly (see DeliveryHandler in the gateway; Booking's proto
  // has no equivalent RPC). Kept in _PageState.activeTrip throughout —
  // Delivery has no payment/rating step, so _state never moves to
  // awaitingPayment/completed the way Ride's _onFinishTrip does.

  Future<void> _onPickupParcel() async {
    if (_state == _PageState.acting) return;
    final trip = _activeTrip;
    if (trip == null) return;
    setState(() { _actingLabel = 'Đang xác nhận đã lấy hàng…'; _state = _PageState.acting; });
    try {
      final deliveryStatus = await _activeTripRepo.pickupParcel(trip.tripId);
      if (!mounted) return;
      setState(() {
        _state = _PageState.activeTrip;
        _activeTrip = trip.copyWith(status: 'in_progress', deliveryStatus: deliveryStatus);
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = _friendlyError(e);
      });
    }
  }

  Future<void> _onStartDelivery() async {
    if (_state == _PageState.acting) return;
    final trip = _activeTrip;
    if (trip == null) return;
    setState(() { _actingLabel = 'Đang bắt đầu giao hàng…'; _state = _PageState.acting; });
    try {
      final deliveryStatus = await _activeTripRepo.startDelivery(trip.tripId);
      if (!mounted) return;
      setState(() {
        _state = _PageState.activeTrip;
        _activeTrip = trip.copyWith(deliveryStatus: deliveryStatus);
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = _friendlyError(e);
      });
    }
  }

  Future<void> _onCompleteDelivery() async {
    if (_state == _PageState.acting) return;
    final trip = _activeTrip;
    if (trip == null) return;
    setState(() { _actingLabel = 'Đang hoàn tất giao hàng…'; _state = _PageState.acting; });
    try {
      final deliveryStatus = await _activeTripRepo.completeDelivery(trip.tripId);
      if (!mounted) return;
      setState(() {
        _state = _PageState.activeTrip;
        _activeTrip = trip.copyWith(deliveryStatus: deliveryStatus);
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = _friendlyError(e);
      });
    }
  }

  void _onDeliveryDone() {
    _activeTripRepo.clearActiveTripId();
    _startPolling();
  }

  Future<void> _onFinishTripTapped() async {
    final confirmed = await AppDialog.confirm(
      context,
      title: 'Kết thúc chuyến đi?',
      message: 'Xác nhận bạn đã đưa khách đến điểm đến. Hành động này không thể hoàn tác.',
      confirmLabel: 'Kết thúc',
    );
    if (confirmed) _onFinishTrip();
  }

  Future<void> _onFinishTrip() async {
    if (_state == _PageState.acting) return;
    final trip = _activeTrip;
    if (trip == null) return;
    setState(() { _actingLabel = 'Đang kết thúc chuyến đi…'; _state = _PageState.acting; });
    // Capture metrics once; if the API call fails and driver retries, reuse
    // the same snapshot rather than calling finish() a second time.
    _finalMetrics ??= _metricsEngine.finish();
    final metrics = _finalMetrics!;
    try {
      final result = await _activeTripRepo.finishTrip(
        tripId: trip.tripId,
        pickupAddress: trip.pickupAddress,
        dropoffAddress: trip.dropoffAddress,
        distanceKm: metrics.distanceKm,
        durationMin: metrics.durationMinutes,
        riderId: trip.riderId,
      );
      _metricsEngine.reset();
      _finalMetrics = null;
      // Active trip ID is kept in storage until the rider pays (status = settled).
      if (!mounted) return;
      setState(() {
        _state = _PageState.awaitingPayment;
        _activeTrip = result;
      });
      _startPaymentPolling();
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = _friendlyError(e);
      });
    }
  }

  // ─── Build ───────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    final showSos = _state == _PageState.activeTrip;
    return Scaffold(
      appBar: AppBar(
        title: Text(_appBarTitle()),
        actions: showSos ? const [SosButton(), SizedBox(width: 4)] : null,
      ),
      body: SafeArea(
        child: AnimatedSwitcher(
          duration: const Duration(milliseconds: 320),
          switchInCurve: Curves.easeOut,
          switchOutCurve: Curves.easeIn,
          transitionBuilder: (child, animation) => FadeTransition(
            opacity: animation,
            child: SlideTransition(
              position: Tween<Offset>(
                begin: const Offset(0, 0.04),
                end: Offset.zero,
              ).animate(animation),
              child: child,
            ),
          ),
          child: KeyedSubtree(key: ValueKey(_state), child: _buildBody()),
        ),
      ),
    );
  }

  String _appBarTitle() {
    if (_state == _PageState.activeTrip) {
      return _activeTrip?.isDelivery == true ? 'Đơn giao hàng' : 'Chuyến đang chạy';
    }
    if (_state == _PageState.awaitingPayment) return 'Đang chờ thanh toán';
    if (_state == _PageState.completed) return 'Chuyến đã hoàn thành';
    return 'Nhận chuyến';
  }

  Widget _buildBody() {
    return switch (_state) {
      _PageState.initializing => const AppLoadingView(label: 'Đang khởi động…'),
      _PageState.polling => _PollingView(onRetry: _poll),
      _PageState.offerAvailable => _offer!.isDelivery
          ? DeliveryOfferCard(
              offer: _offer!,
              countdown: _countdownSeconds,
              onAccept: _onAccept,
              onReject: _onReject,
            )
          : _OfferCard(
              offer: _offer!,
              countdown: _countdownSeconds,
              onAccept: _onAccept,
              onReject: _onReject,
            ),
      _PageState.acting => _showAcceptSuccess
          ? const _SuccessFlash(label: 'Đã chấp nhận chuyến!')
          : AppLoadingView(label: _actingLabel),
      _PageState.activeTrip => _activeTrip!.isDelivery
          ? DeliveryExecutionCard(
              trip: _activeTrip!,
              apiClient: widget.apiClient,
              hasArrived: _hasArrived,
              onArrived: () => _onArrived(),
              onPickupParcel: _onPickupParcel,
              onStartDelivery: _onStartDelivery,
              onCompleteDelivery: _onCompleteDelivery,
              onDone: _onDeliveryDone,
            )
          : _TripExecutionCard(
              trip: _activeTrip!,
              apiClient: widget.apiClient,
              hasArrived: _hasArrived,
              onArrived: () => _onArrived(),
              onStartTrip: _onStartTrip,
              onFinishTrip: _onFinishTripTapped,
            ),
      _PageState.awaitingPayment => _AwaitingPaymentCard(trip: _activeTrip!, apiClient: widget.apiClient),
      _PageState.completed => _TripCompletedCard(
          trip: _activeTrip!,
          onSubmitRating: _submitDriverRating,
          onSkip: _onRatingDone,
        ),
      _PageState.error => AppEmptyState.error(
          subtitle: _errorMessage ?? 'Đã xảy ra lỗi',
          mascotAsset: 'mascot_no_connection.png',
          onAction: _initialize,
        ),
    };
  }
}

// ─── Offer widgets ────────────────────────────────────────────────────────────

/// Brief checkmark-and-label confirmation, swapped in for the generic
/// loading spinner right after an offer accept succeeds (see
/// `_TripPageState._showAcceptSuccess`) — a moment worth a beat of
/// celebration rather than looking identical to every other "please wait".
class _SuccessFlash extends StatelessWidget {
  const _SuccessFlash({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          TweenAnimationBuilder<double>(
            tween: Tween(begin: 0, end: 1),
            duration: const Duration(milliseconds: 320),
            curve: Curves.elasticOut,
            builder: (context, t, child) => Transform.scale(scale: t, child: child),
            child: const Icon(Icons.check_circle, size: AppIconSize.xxl, color: AppColors.primary),
          ),
          const SizedBox(height: AppSpacing.lg),
          Text(
            label,
            style: Theme.of(context)
                .textTheme
                .titleMedium
                ?.copyWith(color: AppColors.primary, fontWeight: FontWeight.w700),
          ),
        ],
      ),
    );
  }
}

class _PollingView extends StatelessWidget {
  const _PollingView({required this.onRetry});

  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const MascotImage(
            asset: 'mascot_driver_ready.png',
            size: MascotSize.large,
            animation: MascotAnimation.fade,
            semanticLabel: 'Đang chờ chuyến mới',
          ),
          const SizedBox(height: AppSpacing.lg),
          Text(
            'Đang chờ chuyến mới…',
            style: Theme.of(context)
                .textTheme
                .titleMedium
                ?.copyWith(color: AppColors.textSecondary),
          ),
          const SizedBox(height: AppSpacing.sm),
          const SizedBox(
            width: 24,
            height: 24,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
        ],
      ),
    );
  }
}

class _OfferCard extends StatelessWidget {
  const _OfferCard({
    required this.offer,
    required this.countdown,
    required this.onAccept,
    required this.onReject,
  });

  final TripOffer offer;
  final int countdown;
  final VoidCallback onAccept;
  final VoidCallback onReject;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: AppCard(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Yêu cầu chuyến mới', style: theme.textTheme.titleLarge),
                _CountdownBadge(seconds: countdown),
              ],
            ),
            const SizedBox(height: AppSpacing.lg),
            const PassengerInfoCard(),
            const SizedBox(height: AppSpacing.lg),
            _AddressRow(
              icon: Icons.location_on,
              color: AppColors.primary,
              label: 'Điểm đón',
              address: offer.pickupAddress,
            ),
            const SizedBox(height: AppSpacing.md),
            _AddressRow(
              icon: Icons.flag,
              color: AppColors.error,
              label: 'Điểm đến',
              address: offer.dropoffAddress,
            ),
            const SizedBox(height: AppSpacing.md),
            const Row(
              children: [
                _InfoChip(
                  icon: Icons.straighten,
                  label: '—',
                  sublabel: 'Khoảng cách',
                ),
                SizedBox(width: AppSpacing.md),
                _InfoChip(
                  icon: Icons.attach_money,
                  label: '—',
                  sublabel: 'Cước phí dự kiến',
                ),
              ],
            ),
            const SizedBox(height: AppSpacing.xxl),
            Row(
              children: [
                Expanded(child: AppButton.danger(label: 'Từ chối', onPressed: onReject)),
                const SizedBox(width: AppSpacing.md),
                Expanded(child: AppButton.primary(label: 'Chấp nhận', onPressed: onAccept)),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _CountdownBadge extends StatelessWidget {
  const _CountdownBadge({required this.seconds});

  final int seconds;

  @override
  Widget build(BuildContext context) {
    final isUrgent = seconds <= 10;
    return AppStatusChip(
      label: '${seconds}s',
      color: isUrgent ? AppColors.error : AppColors.info,
    );
  }
}

// ─── Trip execution widgets ───────────────────────────────────────────────────

class _TripExecutionCard extends StatelessWidget {
  const _TripExecutionCard({
    required this.trip,
    required this.apiClient,
    required this.hasArrived,
    required this.onArrived,
    required this.onStartTrip,
    required this.onFinishTrip,
  });

  final ActiveTrip trip;
  final ApiClient apiClient;
  final bool hasArrived;
  final VoidCallback onArrived;
  final VoidCallback onStartTrip;
  final VoidCallback onFinishTrip;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _StatusBanner(status: trip.status, hasArrived: hasArrived),
          const SizedBox(height: AppSpacing.lg),
          TripTimeline(
            current: trip.status == 'in_progress'
                ? TripTimelineStage.inTrip
                : TripTimelineStage.pickup,
          ),
          const SizedBox(height: AppSpacing.lg),
          PassengerInfoCard(tripId: trip.tripId, apiClient: apiClient),
          const SizedBox(height: AppSpacing.md),
          AppCard(
            padding: const EdgeInsets.all(AppSpacing.xl),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _AddressRow(
                  icon: Icons.location_on,
                  color: AppColors.primary,
                  label: 'Điểm đón',
                  address: trip.pickupAddress,
                ),
                const SizedBox(height: AppSpacing.md),
                _AddressRow(
                  icon: Icons.flag,
                  color: AppColors.error,
                  label: 'Điểm đến',
                  address: trip.dropoffAddress,
                ),
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: AppSpacing.lg),
                  child: Divider(),
                ),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Cước phí dự kiến',
                      style: theme.textTheme.bodyMedium
                          ?.copyWith(color: AppColors.textSecondary),
                    ),
                    Flexible(
                      child: Text(
                        _fareLabel(trip),
                        textAlign: TextAlign.right,
                        style: theme.textTheme.titleMedium,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: AppSpacing.xxl),
                _ActionButton(
                  status: trip.status,
                  hasArrived: hasArrived,
                  onArrived: onArrived,
                  onStartTrip: onStartTrip,
                  onFinishTrip: onFinishTrip,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  static String _fareLabel(ActiveTrip trip) {
    if (trip.finalFare > 0 && trip.fareCurrency.isNotEmpty) {
      return formatMoney(trip.finalFare, trip.fareCurrency);
    }
    return '—';
  }
}

class _StatusBanner extends StatelessWidget {
  const _StatusBanner({required this.status, required this.hasArrived});

  final String status;
  final bool hasArrived;

  @override
  Widget build(BuildContext context) {
    final (label, color, icon) = _statusInfo();
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.lg,
        vertical: AppSpacing.md,
      ),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: AppRadius.mdAll,
        border: Border.all(color: color.withValues(alpha: 0.4)),
      ),
      child: Row(
        children: [
          Icon(icon, color: color, size: AppIconSize.md),
          const SizedBox(width: AppSpacing.sm),
          Flexible(
            child: Text(
              label,
              overflow: TextOverflow.ellipsis,
              style: Theme.of(context)
                  .textTheme
                  .labelLarge
                  ?.copyWith(color: color, fontWeight: FontWeight.w700),
            ),
          ),
        ],
      ),
    );
  }

  (String, Color, IconData) _statusInfo() {
    if (status == 'in_progress') {
      return ('Đang thực hiện', AppColors.info, Icons.directions_car);
    }
    if (status == 'driver_arrived' || (status == 'driver_assigned' && hasArrived)) {
      return ('Đã đến điểm đón', AppColors.primary, Icons.where_to_vote);
    }
    return ('Đang đến điểm đón', AppColors.warning, Icons.navigation);
  }
}

class _ActionButton extends StatelessWidget {
  const _ActionButton({
    required this.status,
    required this.hasArrived,
    required this.onArrived,
    required this.onStartTrip,
    required this.onFinishTrip,
  });

  final String status;
  final bool hasArrived;
  final VoidCallback onArrived;
  final VoidCallback onStartTrip;
  final VoidCallback onFinishTrip;

  @override
  Widget build(BuildContext context) {
    if (status == 'in_progress') {
      return AppButton.danger(label: 'Kết thúc chuyến đi', onPressed: onFinishTrip);
    }
    if (status == 'driver_arrived' || (status == 'driver_assigned' && hasArrived)) {
      return AppButton.primary(label: 'Bắt đầu chuyến đi', onPressed: onStartTrip);
    }
    // driver_assigned, not yet arrived
    return AppButton.outline(label: 'Tôi đã đến điểm đón', onPressed: onArrived);
  }
}

class _AwaitingPaymentCard extends StatelessWidget {
  const _AwaitingPaymentCard({required this.trip, required this.apiClient});

  final ActiveTrip trip;
  final ApiClient apiClient;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Column(
        children: [
          const SizedBox(height: AppSpacing.xxl),
          const SizedBox(
            width: 40,
            height: 40,
            child: CircularProgressIndicator(strokeWidth: 3.5),
          ),
          const SizedBox(height: AppSpacing.xl),
          Text('Đang chờ thanh toán', style: theme.textTheme.headlineSmall),
          const SizedBox(height: AppSpacing.sm),
          Text(
            'Hành khách đang hoàn tất thanh toán.',
            textAlign: TextAlign.center,
            style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
          ),
          const SizedBox(height: AppSpacing.xl),
          const TripTimeline(current: TripTimelineStage.inTrip),
          const SizedBox(height: AppSpacing.xl),
          PassengerInfoCard(tripId: trip.tripId, apiClient: apiClient),
          const SizedBox(height: AppSpacing.lg),
          AppCard(
            padding: const EdgeInsets.all(AppSpacing.xl),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _AddressRow(
                  icon: Icons.location_on,
                  color: AppColors.primary,
                  label: 'Điểm đón',
                  address: trip.pickupAddress,
                ),
                const SizedBox(height: AppSpacing.md),
                _AddressRow(
                  icon: Icons.flag,
                  color: AppColors.error,
                  label: 'Điểm đến',
                  address: trip.dropoffAddress,
                ),
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: AppSpacing.lg),
                  child: Divider(),
                ),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    // Section 6 (Driver Side) fix: this is the trip's gross
                    // fare — what the rider pays — not the driver's net
                    // income (BRB §7.1 commission tiers mean those are
                    // never the same number). The old label here read "Thu
                    // nhập của bạn" ("your income"), which misstated this
                    // exact figure; the real breakdown card below is where
                    // "Thu nhập thực nhận" belongs, honestly marked
                    // "Đang cập nhật" until the backend exposes commission.
                    Text(
                      'Cước phí chuyến đi',
                      style: theme.textTheme.bodyMedium
                          ?.copyWith(color: AppColors.textSecondary),
                    ),
                    Flexible(
                      child: Text(
                        _fareLabel(trip),
                        textAlign: TextAlign.right,
                        style: theme.textTheme.titleLarge?.copyWith(color: AppColors.primary),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: AppSpacing.lg),
          FareBreakdownWaterfall(
            grossAmountCents: trip.finalFare,
            currency: trip.fareCurrency,
            title: 'Thu nhập dự kiến',
          ),
          const SizedBox(height: AppSpacing.lg),
          Text(
            'Bạn sẽ tự động quay lại hàng đợi nhận chuyến sau khi hành khách thanh toán xong.',
            textAlign: TextAlign.center,
            style: theme.textTheme.bodySmall,
          ),
        ],
      ),
    );
  }

  static String _fareLabel(ActiveTrip trip) {
    if (trip.finalFare > 0 && trip.fareCurrency.isNotEmpty) {
      return formatMoney(trip.finalFare, trip.fareCurrency);
    }
    return '—';
  }
}

class _TripCompletedCard extends StatefulWidget {
  const _TripCompletedCard({
    required this.trip,
    required this.onSubmitRating,
    required this.onSkip,
  });

  final ActiveTrip trip;
  final Future<void> Function(int stars, String? comment) onSubmitRating;
  final VoidCallback onSkip;

  @override
  State<_TripCompletedCard> createState() => _TripCompletedCardState();
}

class _TripCompletedCardState extends State<_TripCompletedCard> {
  int _stars = 0;
  bool _submitted = false;
  bool _submitting = false;
  bool _justSucceeded = false;
  String? _error;
  final _commentController = TextEditingController();

  @override
  void dispose() {
    _commentController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_stars == 0 || _submitting) return;
    setState(() { _submitting = true; _error = null; });
    try {
      final comment = _commentController.text.trim();
      await widget.onSubmitRating(_stars, comment.isEmpty ? null : comment);
      if (!mounted) return;
      setState(() { _submitting = false; _justSucceeded = true; });
      await Future.delayed(const Duration(milliseconds: 500));
      if (mounted) setState(() => _submitted = true);
    } catch (_) {
      if (mounted) setState(() { _submitting = false; _error = 'Không thể gửi đánh giá. Bạn có thể bỏ qua.'; });
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Column(
        children: [
          const SizedBox(height: AppSpacing.xxl),
          const MascotImage(
            asset: 'mascot_celebration.png',
            size: MascotSize.large,
            animation: MascotAnimation.bounce,
            semanticLabel: 'Chuyến đi đã hoàn thành',
          ),
          const SizedBox(height: AppSpacing.md),
          Text(
            'Chuyến đi đã hoàn thành!',
            style: theme.textTheme.headlineSmall?.copyWith(color: AppColors.primary),
          ),
          const SizedBox(height: AppSpacing.xl),
          const TripTimeline(current: TripTimelineStage.done),
          const SizedBox(height: AppSpacing.xl),
          AppCard(
            padding: const EdgeInsets.all(AppSpacing.xl),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _AddressRow(
                  icon: Icons.location_on,
                  color: AppColors.primary,
                  label: 'Điểm đón',
                  address: widget.trip.pickupAddress,
                ),
                const SizedBox(height: AppSpacing.md),
                _AddressRow(
                  icon: Icons.flag,
                  color: AppColors.error,
                  label: 'Điểm đến',
                  address: widget.trip.dropoffAddress,
                ),
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: AppSpacing.lg),
                  child: Divider(),
                ),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Cước phí cuối cùng',
                      style: theme.textTheme.bodyMedium
                          ?.copyWith(color: AppColors.textSecondary),
                    ),
                    Flexible(
                      child: Text(
                        _fareLabel(widget.trip),
                        textAlign: TextAlign.right,
                        style: theme.textTheme.titleLarge?.copyWith(color: AppColors.primary),
                      ),
                    ),
                  ],
                ),
                // Real values echoed back by the same `finish` call that
                // just completed this trip (`POST .../finish` already
                // returns `distance_km`/`duration_min` — previously parsed
                // nowhere in this app). Not shown at all when absent (e.g.
                // this card was rebuilt from a re-fetched trip rather than
                // the original finish response) rather than guessed.
                if (widget.trip.distanceKm != null || widget.trip.durationMin != null) ...[
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: AppSpacing.sm),
                    child: Divider(),
                  ),
                  if (widget.trip.distanceKm != null)
                    _SummaryLine(label: 'Quãng đường', value: '${widget.trip.distanceKm!.toStringAsFixed(1)} km'),
                  if (widget.trip.durationMin != null)
                    _SummaryLine(label: 'Thời gian', value: '${widget.trip.durationMin!.round()} phút'),
                ],
              ],
            ),
          ),
          const SizedBox(height: AppSpacing.lg),
          FareBreakdownWaterfall(
            grossAmountCents: widget.trip.finalFare,
            currency: widget.trip.fareCurrency,
          ),
          const SizedBox(height: AppSpacing.xxl),
          if (_submitted) ...[
            const MascotImage(
              asset: 'mascot_rating_5star.png',
              size: MascotSize.medium,
              animation: MascotAnimation.bounce,
              semanticLabel: 'Cảm ơn bạn đã đánh giá',
            ),
            const SizedBox(height: AppSpacing.md),
            Text(
              'Cảm ơn đánh giá của bạn!',
              style: theme.textTheme.bodyMedium?.copyWith(color: AppColors.textSecondary),
            ),
            const SizedBox(height: AppSpacing.lg),
            AppButton.primary(label: 'Quay lại hàng đợi', onPressed: widget.onSkip),
          ] else ...[
            Text('Đánh giá hành khách', style: theme.textTheme.titleMedium),
            const SizedBox(height: AppSpacing.md),
            _StarRow(selected: _stars, onSelect: (s) => setState(() => _stars = s)),
            const SizedBox(height: AppSpacing.md),
            TextField(
              controller: _commentController,
              maxLines: 2,
              maxLength: 200,
              decoration: const InputDecoration(
                hintText: 'Nhận xét (không bắt buộc)…',
              ),
            ),
            if (_error != null) ...[
              const SizedBox(height: AppSpacing.xs),
              Text(_error!, style: const TextStyle(color: AppColors.error, fontSize: 13)),
            ],
            const SizedBox(height: AppSpacing.sm),
            AppButton.primary(
              label: 'Gửi đánh giá',
              isLoading: _submitting,
              isSuccess: _justSucceeded,
              onPressed: _stars > 0 ? _submit : null,
            ),
            if (!_submitting && !_justSucceeded) ...[
              const SizedBox(height: AppSpacing.sm),
              AppButton.text(label: 'Bỏ qua', onPressed: widget.onSkip),
            ],
          ],
        ],
      ),
    );
  }

  static String _fareLabel(ActiveTrip trip) {
    if (trip.finalFare > 0 && trip.fareCurrency.isNotEmpty) {
      return formatMoney(trip.finalFare, trip.fareCurrency);
    }
    return '—';
  }
}

class _StarRow extends StatelessWidget {
  const _StarRow({required this.selected, required this.onSelect});

  final int selected;
  final void Function(int) onSelect;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: List.generate(5, (i) {
        final star = i + 1;
        return IconButton(
          onPressed: () => onSelect(star),
          tooltip: '$star sao',
          icon: Icon(
            star <= selected ? Icons.star : Icons.star_border,
            size: 36,
            color: star <= selected ? AppColors.warning : AppColors.textTertiary,
          ),
        );
      }),
    );
  }
}

// ─── Shared widgets ───────────────────────────────────────────────────────────

class _AddressRow extends StatelessWidget {
  const _AddressRow({
    required this.icon,
    required this.color,
    required this.label,
    required this.address,
  });

  final IconData icon;
  final Color color;
  final String label;
  final String address;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, color: color, size: AppIconSize.md),
        const SizedBox(width: AppSpacing.sm),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label, style: Theme.of(context).textTheme.labelSmall),
              Text(address, style: Theme.of(context).textTheme.bodyMedium),
            ],
          ),
        ),
      ],
    );
  }
}

class _SummaryLine extends StatelessWidget {
  const _SummaryLine({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs / 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: theme.textTheme.bodySmall?.copyWith(color: AppColors.textSecondary)),
          Flexible(
            child: Text(
              value,
              textAlign: TextAlign.right,
              style: theme.textTheme.bodySmall?.copyWith(fontWeight: FontWeight.w600),
            ),
          ),
        ],
      ),
    );
  }
}

class _InfoChip extends StatelessWidget {
  const _InfoChip({
    required this.icon,
    required this.label,
    required this.sublabel,
  });

  final IconData icon;
  final String label;
  final String sublabel;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.md,
          vertical: AppSpacing.sm,
        ),
        decoration: BoxDecoration(
          color: AppColors.surfaceAlt,
          borderRadius: AppRadius.smAll,
        ),
        child: Row(
          children: [
            Icon(icon, size: AppIconSize.sm, color: AppColors.textSecondary),
            const SizedBox(width: 6),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(label, style: Theme.of(context).textTheme.titleSmall),
                Text(sublabel, style: Theme.of(context).textTheme.labelSmall),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
