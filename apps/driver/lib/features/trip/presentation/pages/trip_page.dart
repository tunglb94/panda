import 'dart:async';

import 'package:flutter/material.dart';

import '../../../../core/location/location_engine.dart';
import '../../../../core/network/api_client.dart';
import '../../../../core/trip_metrics/trip_metrics.dart';
import '../../../../core/trip_metrics/trip_metrics_engine.dart';
import '../../data/active_trip_repository.dart';
import '../../data/trip_offer_repository.dart';

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

class _TripPageState extends State<TripPage> {
  late final TripOfferRepository _offerRepo;
  late final ActiveTripRepository _activeTripRepo;
  late final TripMetricsEngine _metricsEngine;

  _PageState _state = _PageState.initializing;
  TripOffer? _offer;
  ActiveTrip? _activeTrip;
  String? _errorMessage;
  int _countdownSeconds = 0;
  bool _hasArrived = false;
  TripMetrics? _finalMetrics;

  Timer? _pollTimer;
  Timer? _countdownTimer;
  Timer? _paymentPollTimer;
  bool _isPollingActive = false;

  @override
  void initState() {
    super.initState();
    _offerRepo = TripOfferRepository(apiClient: widget.apiClient);
    _activeTripRepo = ActiveTripRepository(apiClient: widget.apiClient);
    _metricsEngine = TripMetricsEngine(locationStream: widget.locationStream);
    _initialize();
  }

  @override
  void dispose() {
    _pollTimer?.cancel();
    _countdownTimer?.cancel();
    _paymentPollTimer?.cancel();
    _metricsEngine.reset();
    super.dispose();
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
            _hasArrived = false;
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
            _errorMessage = e.message;
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
        _state == _PageState.initializing) return;

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
        }
      }
    } on ApiException catch (e) {
      if (!mounted) return;
      if (_state != _PageState.error) {
        setState(() {
          _state = _PageState.error;
          _errorMessage = e.message;
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
    final trip = _activeTrip;
    if (trip == null) return;
    try {
      final updated = await _activeTripRepo.fetchTrip(trip.tripId);
      if (!mounted) return;
      if (updated.status == 'settled') {
        _paymentPollTimer?.cancel();
        _paymentPollTimer = null;
        await _activeTripRepo.clearActiveTripId();
        if (!mounted) return;
        setState(() {
          _state = _PageState.completed;
          _activeTrip = updated;
        });
        Future.delayed(const Duration(seconds: 4), () {
          if (mounted && _state == _PageState.completed) {
            _startPolling();
          }
        });
      } else {
        setState(() => _activeTrip = updated);
      }
    } on ApiException catch (_) {
      // Ignore transient poll errors; rider may still be paying.
    }
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
    final offer = _offer;
    if (offer == null) return;
    _pollTimer?.cancel();
    _countdownTimer?.cancel();
    setState(() => _state = _PageState.acting);
    try {
      await _offerRepo.acceptOffer(offer.tripId);
      await _activeTripRepo.saveActiveTripId(offer.tripId);
      if (!mounted) return;
      setState(() {
        _state = _PageState.activeTrip;
        _activeTrip = ActiveTrip(
          tripId: offer.tripId,
          pickupAddress: offer.pickupAddress,
          dropoffAddress: offer.dropoffAddress,
          status: 'driver_assigned',
        );
        _hasArrived = false;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = e.message;
      });
    }
  }

  Future<void> _onReject() async {
    final offer = _offer;
    if (offer == null) return;
    _pollTimer?.cancel();
    _countdownTimer?.cancel();
    setState(() => _state = _PageState.acting);
    try {
      await _offerRepo.rejectOffer(offer.tripId);
      if (!mounted) return;
      _startPolling();
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _state = _PageState.error;
        _errorMessage = e.message;
      });
    }
  }

  // ─── Trip execution actions ──────────────────────────────────────────────────

  void _onArrived() {
    setState(() => _hasArrived = true);
  }

  Future<void> _onStartTrip() async {
    final trip = _activeTrip;
    if (trip == null) return;
    setState(() => _state = _PageState.acting);
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
        _errorMessage = e.message;
      });
    }
  }

  Future<void> _onFinishTrip() async {
    final trip = _activeTrip;
    if (trip == null) return;
    setState(() => _state = _PageState.acting);
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
        _errorMessage = e.message;
      });
    }
  }

  // ─── Build ───────────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(_appBarTitle())),
      body: SafeArea(child: _buildBody()),
    );
  }

  String _appBarTitle() {
    if (_state == _PageState.activeTrip) return 'Active Trip';
    if (_state == _PageState.awaitingPayment) return 'Awaiting Payment';
    if (_state == _PageState.completed) return 'Trip Completed';
    return 'Trip Offers';
  }

  Widget _buildBody() {
    return switch (_state) {
      _PageState.initializing =>
        const Center(child: CircularProgressIndicator()),
      _PageState.polling => _PollingView(onRetry: _poll),
      _PageState.offerAvailable => _OfferCard(
          offer: _offer!,
          countdown: _countdownSeconds,
          onAccept: _onAccept,
          onReject: _onReject,
        ),
      _PageState.acting => const Center(child: CircularProgressIndicator()),
      _PageState.activeTrip => _TripExecutionCard(
          trip: _activeTrip!,
          hasArrived: _hasArrived,
          onArrived: _onArrived,
          onStartTrip: _onStartTrip,
          onFinishTrip: _onFinishTrip,
        ),
      _PageState.awaitingPayment => _AwaitingPaymentCard(trip: _activeTrip!),
      _PageState.completed => _TripCompletedCard(trip: _activeTrip!),
      _PageState.error => _ErrorView(
          message: _errorMessage ?? 'An error occurred',
          onRetry: _initialize,
        ),
    };
  }
}

// ─── Offer widgets ────────────────────────────────────────────────────────────

class _PollingView extends StatelessWidget {
  const _PollingView({required this.onRetry});

  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.directions_car_outlined, size: 64, color: cs.outline),
          const SizedBox(height: 16),
          Text(
            'Waiting for trip offers…',
            style: Theme.of(context)
                .textTheme
                .titleMedium
                ?.copyWith(color: cs.onSurfaceVariant),
          ),
          const SizedBox(height: 8),
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
    final cs = theme.colorScheme;
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'New Trip Request',
                    style: theme.textTheme.titleLarge
                        ?.copyWith(fontWeight: FontWeight.bold),
                  ),
                  _CountdownBadge(seconds: countdown),
                ],
              ),
              const SizedBox(height: 20),
              _AddressRow(
                icon: Icons.location_on,
                color: cs.primary,
                label: 'Pickup',
                address: offer.pickupAddress,
              ),
              const SizedBox(height: 12),
              _AddressRow(
                icon: Icons.flag,
                color: cs.error,
                label: 'Destination',
                address: offer.dropoffAddress,
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  _InfoChip(
                    icon: Icons.straighten,
                    label: '—',
                    sublabel: 'Distance',
                  ),
                  const SizedBox(width: 12),
                  _InfoChip(
                    icon: Icons.attach_money,
                    label: '—',
                    sublabel: 'Est. fare',
                  ),
                ],
              ),
              const SizedBox(height: 24),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: onReject,
                      style: OutlinedButton.styleFrom(
                        foregroundColor: cs.error,
                        side: BorderSide(color: cs.error),
                        padding: const EdgeInsets.symmetric(vertical: 16),
                      ),
                      child: const Text('Reject'),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: FilledButton(
                      onPressed: onAccept,
                      style: FilledButton.styleFrom(
                        padding: const EdgeInsets.symmetric(vertical: 16),
                      ),
                      child: const Text('Accept'),
                    ),
                  ),
                ],
              ),
            ],
          ),
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
    final cs = Theme.of(context).colorScheme;
    final isUrgent = seconds <= 10;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: isUrgent ? cs.errorContainer : cs.secondaryContainer,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        '${seconds}s',
        style: Theme.of(context).textTheme.labelLarge?.copyWith(
              color: isUrgent ? cs.onErrorContainer : cs.onSecondaryContainer,
              fontWeight: FontWeight.bold,
            ),
      ),
    );
  }
}

// ─── Trip execution widgets ───────────────────────────────────────────────────

class _TripExecutionCard extends StatelessWidget {
  const _TripExecutionCard({
    required this.trip,
    required this.hasArrived,
    required this.onArrived,
    required this.onStartTrip,
    required this.onFinishTrip,
  });

  final ActiveTrip trip;
  final bool hasArrived;
  final VoidCallback onArrived;
  final VoidCallback onStartTrip;
  final VoidCallback onFinishTrip;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _StatusBanner(status: trip.status, hasArrived: hasArrived),
          const SizedBox(height: 12),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _AddressRow(
                    icon: Icons.location_on,
                    color: cs.primary,
                    label: 'Pickup',
                    address: trip.pickupAddress,
                  ),
                  const SizedBox(height: 12),
                  _AddressRow(
                    icon: Icons.flag,
                    color: cs.error,
                    label: 'Destination',
                    address: trip.dropoffAddress,
                  ),
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Divider(),
                  ),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Estimated fare',
                        style: theme.textTheme.bodyMedium
                            ?.copyWith(color: cs.onSurfaceVariant),
                      ),
                      Text(
                        _fareLabel(trip),
                        style: theme.textTheme.titleMedium
                            ?.copyWith(fontWeight: FontWeight.bold),
                      ),
                    ],
                  ),
                  const SizedBox(height: 24),
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
          ),
        ],
      ),
    );
  }

  static String _fareLabel(ActiveTrip trip) {
    if (trip.finalFare > 0 && trip.fareCurrency.isNotEmpty) {
      final amount = trip.finalFare / 100;
      return '${trip.fareCurrency} ${amount.toStringAsFixed(2)}';
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
    final cs = Theme.of(context).colorScheme;
    final (label, color, icon) = _statusInfo(cs);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withValues(alpha: 0.4)),
      ),
      child: Row(
        children: [
          Icon(icon, color: color, size: 20),
          const SizedBox(width: 10),
          Text(
            label,
            style: Theme.of(context)
                .textTheme
                .labelLarge
                ?.copyWith(color: color, fontWeight: FontWeight.w600),
          ),
        ],
      ),
    );
  }

  (String, Color, IconData) _statusInfo(ColorScheme cs) {
    if (status == 'in_progress') {
      return ('In Progress', cs.tertiary, Icons.directions_car);
    }
    if (status == 'driver_assigned' && hasArrived) {
      return ('Arrived at Pickup', cs.primary, Icons.where_to_vote);
    }
    return ('Heading to Pickup', cs.secondary, Icons.navigation);
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
    final cs = Theme.of(context).colorScheme;
    if (status == 'in_progress') {
      return FilledButton(
        onPressed: onFinishTrip,
        style: FilledButton.styleFrom(
          backgroundColor: cs.errorContainer,
          foregroundColor: cs.onErrorContainer,
          minimumSize: const Size.fromHeight(52),
        ),
        child: const Text('Complete Trip'),
      );
    }
    if (status == 'driver_assigned' && hasArrived) {
      return FilledButton(
        onPressed: onStartTrip,
        style: FilledButton.styleFrom(
          minimumSize: const Size.fromHeight(52),
        ),
        child: const Text('Start Trip'),
      );
    }
    // driver_assigned, not yet arrived
    return OutlinedButton(
      onPressed: onArrived,
      style: OutlinedButton.styleFrom(
        minimumSize: const Size.fromHeight(52),
      ),
      child: const Text('I\'ve Arrived at Pickup'),
    );
  }
}

class _AwaitingPaymentCard extends StatelessWidget {
  const _AwaitingPaymentCard({required this.trip});

  final ActiveTrip trip;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          const SizedBox(height: 24),
          SizedBox(
            width: 64,
            height: 64,
            child: CircularProgressIndicator(
              strokeWidth: 5,
              color: cs.primary,
            ),
          ),
          const SizedBox(height: 20),
          Text(
            'Waiting for Payment',
            style: theme.textTheme.headlineSmall
                ?.copyWith(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          Text(
            'The rider is completing payment.',
            style: theme.textTheme.bodyMedium
                ?.copyWith(color: cs.onSurfaceVariant),
          ),
          const SizedBox(height: 24),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _AddressRow(
                    icon: Icons.location_on,
                    color: cs.primary,
                    label: 'Pickup',
                    address: trip.pickupAddress,
                  ),
                  const SizedBox(height: 12),
                  _AddressRow(
                    icon: Icons.flag,
                    color: cs.error,
                    label: 'Destination',
                    address: trip.dropoffAddress,
                  ),
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Divider(),
                  ),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Your earnings',
                        style: theme.textTheme.bodyMedium
                            ?.copyWith(color: cs.onSurfaceVariant),
                      ),
                      Text(
                        _fareLabel(trip),
                        style: theme.textTheme.titleLarge?.copyWith(
                          fontWeight: FontWeight.bold,
                          color: cs.primary,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          Text(
            'You will automatically return to the offer queue once paid.',
            textAlign: TextAlign.center,
            style: theme.textTheme.bodySmall
                ?.copyWith(color: cs.onSurfaceVariant),
          ),
        ],
      ),
    );
  }

  static String _fareLabel(ActiveTrip trip) {
    if (trip.finalFare > 0 && trip.fareCurrency.isNotEmpty) {
      final amount = trip.finalFare / 100;
      return '${trip.fareCurrency} ${amount.toStringAsFixed(2)}';
    }
    return '—';
  }
}

class _TripCompletedCard extends StatelessWidget {
  const _TripCompletedCard({required this.trip});

  final ActiveTrip trip;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          const SizedBox(height: 24),
          Icon(Icons.check_circle_outline, size: 72, color: cs.primary),
          const SizedBox(height: 12),
          Text(
            'Trip Completed!',
            style: theme.textTheme.headlineSmall
                ?.copyWith(color: cs.primary, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 24),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _AddressRow(
                    icon: Icons.location_on,
                    color: cs.primary,
                    label: 'Pickup',
                    address: trip.pickupAddress,
                  ),
                  const SizedBox(height: 12),
                  _AddressRow(
                    icon: Icons.flag,
                    color: cs.error,
                    label: 'Destination',
                    address: trip.dropoffAddress,
                  ),
                  const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Divider(),
                  ),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Final fare',
                        style: theme.textTheme.bodyMedium
                            ?.copyWith(color: cs.onSurfaceVariant),
                      ),
                      Text(
                        _fareLabel(trip),
                        style: theme.textTheme.titleLarge?.copyWith(
                          fontWeight: FontWeight.bold,
                          color: cs.primary,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          Text(
            'Returning to offer queue…',
            style: theme.textTheme.bodySmall
                ?.copyWith(color: cs.onSurfaceVariant),
          ),
        ],
      ),
    );
  }

  static String _fareLabel(ActiveTrip trip) {
    if (trip.finalFare > 0 && trip.fareCurrency.isNotEmpty) {
      final amount = trip.finalFare / 100;
      return '${trip.fareCurrency} ${amount.toStringAsFixed(2)}';
    }
    return '—';
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
        Icon(icon, color: color, size: 20),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: Theme.of(context).textTheme.labelSmall?.copyWith(
                      color: Theme.of(context).colorScheme.outline,
                    ),
              ),
              Text(address, style: Theme.of(context).textTheme.bodyMedium),
            ],
          ),
        ),
      ],
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
    final cs = Theme.of(context).colorScheme;
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          color: cs.surfaceContainerHighest,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          children: [
            Icon(icon, size: 16, color: cs.onSurfaceVariant),
            const SizedBox(width: 6),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: Theme.of(context)
                      .textTheme
                      .titleSmall
                      ?.copyWith(fontWeight: FontWeight.bold),
                ),
                Text(
                  sublabel,
                  style: Theme.of(context)
                      .textTheme
                      .labelSmall
                      ?.copyWith(color: cs.onSurfaceVariant),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.wifi_off_outlined, size: 48, color: cs.error),
            const SizedBox(height: 16),
            Text(
              message,
              textAlign: TextAlign.center,
              style: Theme.of(context)
                  .textTheme
                  .bodyMedium
                  ?.copyWith(color: cs.onSurfaceVariant),
            ),
            const SizedBox(height: 16),
            FilledButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }
}
