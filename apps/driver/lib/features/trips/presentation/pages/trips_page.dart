import 'package:flutter/material.dart';

import 'package:driver/shared/widgets/async_state_view.dart';

import '../../data/driver_trip_offer_repository.dart';
import '../../domain/models/dispatch_accept_result.dart';
import '../../domain/models/route_progress_model.dart';
import '../../domain/models/trip_offer.dart';
import '../../domain/models/trip_offer_state.dart';
import '../widgets/trip_offer_view.dart';
import 'trip_offer_preview_menu_page.dart';

/// Trips tab: the incoming-trip-offer flow. Fetches a mock offer through
/// `DriverTripOfferRepository` via the shared `AsyncStateView` (Loading/
/// Success/Empty/Error), then drives a local offer-lifecycle state machine
/// (`TripOfferState`: New Offer → Accepting → Assigned → Navigating to
/// Pickup, or → Failed/Timeout, or → Rejected/Expired) once an offer is
/// showing. The two state machines are deliberately kept separate —
/// `AsyncStateView` is untouched.
class TripsPage extends StatefulWidget {
  const TripsPage({super.key});

  @override
  State<TripsPage> createState() => _TripsPageState();
}

class _TripsPageState extends State<TripsPage> {
  static const _repository = DriverTripOfferRepository();

  DriverTripOfferDemoMode _mode = DriverTripOfferDemoMode.normal;
  late Future<TripOffer?> _future;
  TripOfferState _state = TripOfferState.newOffer;

  /// Dev-only control for what `acceptOffer()` should return the next time
  /// Accept is pressed — lets the live page (and tests) exercise the
  /// Failed/Timeout paths without depending on real randomness.
  DispatchAcceptStatus _nextAcceptOutcome = DispatchAcceptStatus.success;

  /// Dev-only control for the traffic level `fetchRouteProgress()` returns
  /// the next time Start Navigation is pressed.
  TrafficLevel _traffic = TrafficLevel.normal;

  Future<RouteProgressModel>? _routeFuture;

  @override
  void initState() {
    super.initState();
    _load();
  }

  void _load() {
    setState(() {
      _state = TripOfferState.newOffer;
      _future = _repository.fetchOffer(mode: _mode);
    });
  }

  Future<void> _handleAccept() async {
    if (_state != TripOfferState.newOffer) return;
    setState(() => _state = TripOfferState.accepting);

    final result = await _repository.acceptOffer(outcome: _nextAcceptOutcome);
    if (!mounted) return;

    setState(() {
      _state = switch (result.status) {
        DispatchAcceptStatus.success => TripOfferState.assigned,
        DispatchAcceptStatus.failed => TripOfferState.failed,
        DispatchAcceptStatus.timeout => TripOfferState.timeout,
      };
    });
  }

  void _handleReject() {
    if (_state != TripOfferState.newOffer) return;
    setState(() => _state = TripOfferState.rejected);
  }

  void _handleExpired() {
    // Guard against the countdown firing while its own content is already
    // fading out of an `AnimatedSwitcher` transition (e.g. Accept was
    // pressed a moment earlier) — only a countdown that is still genuinely
    // showing may expire the offer.
    if (_state != TripOfferState.newOffer) return;
    setState(() => _state = TripOfferState.expired);
  }

  void _handleRetry() {
    setState(() => _state = TripOfferState.newOffer);
  }

  void _handleNavigate() {
    if (_state != TripOfferState.assigned) return;
    setState(() {
      _state = TripOfferState.navigatingToPickup;
      _routeFuture = _repository.fetchRouteProgress(traffic: _traffic);
    });
  }

  void _handleArrived() {
    // Not a hard timer — this is only called by the route-progress ticker
    // once its mock progress genuinely reaches 0%.
    if (_state != TripOfferState.navigatingToPickup) return;
    setState(() => _state = TripOfferState.arrivedAtPickup);
  }

  void _handlePassengerOnBoard() {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Passenger On Board is a placeholder — not yet implemented.')),
    );
  }

  void _handleContactRider() {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Contact Rider is a placeholder — not yet implemented.')),
    );
  }

  void _handleCancelTrip() {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Cancel Trip is a placeholder — not yet implemented.')),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Trips'),
        actions: [
          PopupMenuButton<DriverTripOfferDemoMode>(
            tooltip: 'Preview state (dev)',
            icon: const Icon(Icons.tune),
            onSelected: (mode) {
              _mode = mode;
              _load();
            },
            itemBuilder: (context) => const [
              PopupMenuItem(value: DriverTripOfferDemoMode.normal, child: Text('Normal')),
              PopupMenuItem(value: DriverTripOfferDemoMode.empty, child: Text('Empty (dev)')),
              PopupMenuItem(value: DriverTripOfferDemoMode.error, child: Text('Error (dev)')),
            ],
          ),
          PopupMenuButton<DispatchAcceptStatus>(
            tooltip: 'Accept outcome (dev)',
            icon: const Icon(Icons.rule),
            onSelected: (outcome) => _nextAcceptOutcome = outcome,
            itemBuilder: (context) => const [
              PopupMenuItem(value: DispatchAcceptStatus.success, child: Text('Accept → Success')),
              PopupMenuItem(value: DispatchAcceptStatus.failed, child: Text('Accept → Failed (dev)')),
              PopupMenuItem(
                  value: DispatchAcceptStatus.timeout, child: Text('Accept → Timeout (dev)')),
            ],
          ),
          PopupMenuButton<TrafficLevel>(
            tooltip: 'Traffic (dev)',
            icon: const Icon(Icons.traffic),
            onSelected: (traffic) => _traffic = traffic,
            itemBuilder: (context) => [
              for (final level in TrafficLevel.values) PopupMenuItem(value: level, child: Text(level.label)),
            ],
          ),
          IconButton(
            tooltip: 'Trip offer states (dev)',
            icon: const Icon(Icons.fact_check_outlined),
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute(builder: (_) => const TripOfferPreviewMenuPage()),
            ),
          ),
        ],
      ),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: AsyncStateView<TripOffer?>(
                future: _future,
                isEmpty: (offer) => offer == null,
                emptyBuilder: (context) => Padding(
                  padding: const EdgeInsets.symmetric(vertical: 48, horizontal: 24),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.local_taxi_outlined, size: 48, color: Colors.grey.shade400),
                      const SizedBox(height: 12),
                      const Text('No incoming trip offers right now',
                          style: TextStyle(fontWeight: FontWeight.w600)),
                      const SizedBox(height: 4),
                      Text(
                        "You'll see new offers here as soon as they come in.",
                        textAlign: TextAlign.center,
                        style: TextStyle(color: Colors.grey.shade500),
                      ),
                    ],
                  ),
                ),
                errorBuilder: (context, error) => Padding(
                  padding: const EdgeInsets.symmetric(vertical: 48, horizontal: 24),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.error_outline, size: 48, color: Colors.red.shade400),
                      const SizedBox(height: 12),
                      const Text("Couldn't load trip offers",
                          style: TextStyle(fontWeight: FontWeight.w600)),
                      const SizedBox(height: 12),
                      OutlinedButton(onPressed: _load, child: const Text('Retry')),
                    ],
                  ),
                ),
                successBuilder: (context, offer) => TripOfferView(
                  state: _state,
                  offer: offer!,
                  onAccept: _handleAccept,
                  onReject: _handleReject,
                  onExpired: _handleExpired,
                  onNavigate: _handleNavigate,
                  onRetry: _handleRetry,
                  routeFuture: _routeFuture,
                  onArrived: _handleArrived,
                  onContactRider: _handleContactRider,
                  onCancelTrip: _handleCancelTrip,
                  onPassengerOnBoard: _handlePassengerOnBoard,
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
