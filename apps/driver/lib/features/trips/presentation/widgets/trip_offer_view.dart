import 'package:flutter/material.dart';

import 'package:driver/shared/widgets/async_state_view.dart';

import '../../domain/models/route_progress_model.dart';
import '../../domain/models/trip_offer.dart';
import '../../domain/models/trip_offer_state.dart';
import 'accept_loading_button.dart';
import 'countdown_indicator.dart';
import 'dispatch_status_banner.dart';
import 'driver_arrival_card.dart';
import 'driver_navigation_card.dart';
import 'fare_estimate_card.dart';
import 'trip_action_buttons.dart';
import 'trip_assigned_card.dart';
import 'trip_offer_card.dart';

/// Renders the correct content for [state] — the New Offer card (with
/// countdown + actions), the Accepting spinner, the Assigned screen, the
/// Navigating-to-Pickup screen, the Arrived-at-Pickup screen, a Failed/
/// Timeout retry banner, or a simple result banner for Rejected/Expired.
/// Content-only (no `Scaffold`), reused by both the live `TripsPage` flow
/// and the dev preview pages, cross-fading between states.
class TripOfferView extends StatelessWidget {
  const TripOfferView({
    super.key,
    required this.state,
    required this.offer,
    this.onAccept,
    this.onReject,
    this.onExpired,
    this.onNavigate,
    this.onRetry,
    this.routeFuture,
    this.onArrived,
    this.onContactRider,
    this.onCancelTrip,
    this.onPassengerOnBoard,
  });

  final TripOfferState state;
  final TripOffer offer;
  final VoidCallback? onAccept;
  final VoidCallback? onReject;
  final VoidCallback? onExpired;
  final VoidCallback? onNavigate;
  final VoidCallback? onRetry;

  /// Fetched once when entering `navigatingToPickup` — see `TripsPage`.
  final Future<RouteProgressModel>? routeFuture;

  /// Fired once by the route-progress ticker when it reaches 0% — drives
  /// `navigatingToPickup -> arrivedAtPickup` (Phase D-06). Not a hard timer:
  /// the transition happens exactly when the mock progress completes.
  final VoidCallback? onArrived;
  final VoidCallback? onContactRider;
  final VoidCallback? onCancelTrip;
  final VoidCallback? onPassengerOnBoard;

  @override
  Widget build(BuildContext context) {
    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 350),
      child: KeyedSubtree(
        key: ValueKey(state),
        child: switch (state) {
          TripOfferState.newOffer => _NewOfferContent(
              offer: offer,
              onAccept: onAccept ?? () {},
              onReject: onReject ?? () {},
              onExpired: onExpired ?? () {},
            ),
          TripOfferState.accepting => _AcceptingContent(offer: offer),
          TripOfferState.assigned => TripAssignedCard(
              offer: offer,
              onNavigate: onNavigate ?? () {},
            ),
          TripOfferState.navigatingToPickup => _NavigatingToPickupContent(
              offer: offer,
              routeFuture: routeFuture ??
                  Future.value(
                    RouteProgressModel.mock(progress: 100, trafficLevel: TrafficLevel.normal),
                  ),
              onArrived: onArrived ?? () {},
              onContactRider: onContactRider ?? () {},
              onCancelTrip: onCancelTrip ?? () {},
            ),
          TripOfferState.arrivedAtPickup => DriverArrivalCard(
              offer: offer,
              onPassengerOnBoard: onPassengerOnBoard ?? () {},
              onContactRider: onContactRider ?? () {},
              onCancelTrip: onCancelTrip ?? () {},
            ),
          TripOfferState.rejected => _ResultBanner(
              icon: Icons.cancel,
              color: Colors.red.shade600,
              title: 'Trip Rejected',
              message: 'This offer has been declined.',
            ),
          TripOfferState.expired => _ResultBanner(
              icon: Icons.timer_off_outlined,
              color: Colors.grey.shade600,
              title: 'Offer Expired',
              message: "You didn't respond in time — the offer has been "
                  'reassigned to another driver.',
            ),
          TripOfferState.failed => DispatchStatusBanner(
              icon: Icons.error_outline,
              color: Colors.red.shade600,
              title: 'Unable to accept trip.',
              message: 'Try again.',
              onRetry: onRetry ?? () {},
            ),
          TripOfferState.timeout => DispatchStatusBanner(
              icon: Icons.timer_off_outlined,
              color: Colors.orange.shade700,
              title: 'Dispatch timeout.',
              message: 'Please retry.',
              onRetry: onRetry ?? () {},
            ),
        },
      ),
    );
  }
}

class _NewOfferContent extends StatelessWidget {
  const _NewOfferContent({
    required this.offer,
    required this.onAccept,
    required this.onReject,
    required this.onExpired,
  });

  final TripOffer offer;
  final VoidCallback onAccept;
  final VoidCallback onReject;
  final VoidCallback onExpired;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Center(child: CountdownIndicator(onExpired: onExpired)),
        const SizedBox(height: 16),
        TripOfferCard(offer: offer),
        const SizedBox(height: 16),
        FareEstimateCard(offer: offer),
        const SizedBox(height: 20),
        TripActionButtons(onAccept: onAccept, onReject: onReject),
      ],
    );
  }
}

/// Shown the instant Accept is pressed. Deliberately has **no**
/// `CountdownIndicator` — removing it from the tree (via the parent
/// `AnimatedSwitcher` swapping to this content) is what stops the countdown;
/// see `TripsPage._handleExpired` for the belt-and-suspenders guard against
/// a stray `onExpired` firing during the swap's cross-fade.
class _AcceptingContent extends StatelessWidget {
  const _AcceptingContent({required this.offer});

  final TripOffer offer;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TripOfferCard(offer: offer),
        const SizedBox(height: 16),
        FareEstimateCard(offer: offer),
        const SizedBox(height: 20),
        const AcceptLoadingButton(),
      ],
    );
  }
}

/// Wraps the `navigatingToPickup` case's own async fetch (its own nested
/// `AsyncStateView<RouteProgressModel>`) — kept separate from the outer
/// `AsyncStateView<TripOffer?>` that gates the whole page, per this
/// project's "never merge AsyncState with the business state machine" rule.
class _NavigatingToPickupContent extends StatelessWidget {
  const _NavigatingToPickupContent({
    required this.offer,
    required this.routeFuture,
    required this.onArrived,
    required this.onContactRider,
    required this.onCancelTrip,
  });

  final TripOffer offer;
  final Future<RouteProgressModel> routeFuture;
  final VoidCallback onArrived;
  final VoidCallback onContactRider;
  final VoidCallback onCancelTrip;

  @override
  Widget build(BuildContext context) {
    return AsyncStateView<RouteProgressModel>(
      future: routeFuture,
      successBuilder: (context, route) => _RouteProgressTicker(
        offer: offer,
        initialRoute: route,
        onArrived: onArrived,
        onContactRider: onContactRider,
        onCancelTrip: onCancelTrip,
      ),
    );
  }
}

/// Ticks the mock route progress down (100 → 80 → 60 → 40 → 20 → 0, every
/// 2s) purely locally, without any further repository calls. Once it
/// reaches 0 ("Arrived"), it fires [onArrived] exactly once and stops
/// scheduling — no hard timer drives the transition, the progress reaching
/// its end does. Stopping means it's also safe to `pumpAndSettle()` past it
/// in tests (unlike the perpetually re-armed offer countdown).
class _RouteProgressTicker extends StatefulWidget {
  const _RouteProgressTicker({
    required this.offer,
    required this.initialRoute,
    required this.onArrived,
    required this.onContactRider,
    required this.onCancelTrip,
  });

  final TripOffer offer;
  final RouteProgressModel initialRoute;
  final VoidCallback onArrived;
  final VoidCallback onContactRider;
  final VoidCallback onCancelTrip;

  @override
  State<_RouteProgressTicker> createState() => _RouteProgressTickerState();
}

class _RouteProgressTickerState extends State<_RouteProgressTicker> {
  late RouteProgressModel _route;
  bool _stopped = false;

  @override
  void initState() {
    super.initState();
    _route = widget.initialRoute;
    if (_route.progress <= 0) {
      // Defensive edge case (never hit from the live repository, which
      // always seeds at 100) — defer past this build so we don't call
      // setState-in-disguise (via onArrived) while a parent is mid-build.
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!_stopped && mounted) widget.onArrived();
      });
    } else {
      _scheduleTick();
    }
  }

  Future<void> _scheduleTick() async {
    await Future.delayed(const Duration(seconds: 2));
    if (_stopped || !mounted) return;
    final next = _route.stepDown();
    setState(() => _route = next);
    if (next.progress <= 0) {
      widget.onArrived();
      return;
    }
    _scheduleTick();
  }

  @override
  void dispose() {
    _stopped = true;
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return DriverNavigationCard(
      offer: widget.offer,
      route: _route,
      onContactRider: widget.onContactRider,
      onCancelTrip: widget.onCancelTrip,
    );
  }
}

class _ResultBanner extends StatelessWidget {
  const _ResultBanner({
    required this.icon,
    required this.color,
    required this.title,
    required this.message,
  });

  final IconData icon;
  final Color color;
  final String title;
  final String message;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 48, horizontal: 16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          TweenAnimationBuilder<double>(
            tween: Tween(begin: 0.6, end: 1.0),
            duration: const Duration(milliseconds: 350),
            curve: Curves.easeOutBack,
            builder: (context, scale, child) => Transform.scale(scale: scale, child: child),
            child: Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(color: color.withValues(alpha: 0.12), shape: BoxShape.circle),
              child: Icon(icon, size: 44, color: color),
            ),
          ),
          const SizedBox(height: 20),
          Text(title, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          Text(
            message,
            textAlign: TextAlign.center,
            style: TextStyle(color: Colors.grey.shade600),
          ),
        ],
      ),
    );
  }
}
