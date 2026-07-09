import 'dart:async';

import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';
import 'package:rider/features/trip/data/trip_repository.dart';
import 'package:rider/features/trip/domain/models/mock_trip_catalog.dart';
import 'package:rider/features/trip/domain/models/rider_trip_status.dart';

import 'driver_arriving_view.dart';
import 'driver_assigned_view.dart';
import 'searching_driver_view.dart';
import 'trip_cancelled_view.dart';
import 'trip_completed_view.dart';
import 'trip_in_progress_view.dart';

/// Live trip lifecycle screen.
///
/// Polls `GET /api/v1/rides/{tripId}` every 5 seconds and animates through
/// [RiderTripStatus] states. Polling stops automatically on a terminal state
/// (completed or cancelled).
///
/// [onDriverAssigned] is called once when the status first transitions to
/// [RiderTripStatus.driverAssigned] with a non-empty driverId. Callers that
/// have a map view can use this to start live driver tracking.
class TripLifecyclePage extends StatefulWidget {
  const TripLifecyclePage({
    super.key,
    required this.tripId,
    required this.tripSelection,
    required this.apiClient,
    this.onDriverAssigned,
  });

  final String tripId;
  final TripSelection tripSelection;
  final ApiClient apiClient;
  final void Function(String driverId)? onDriverAssigned;

  @override
  State<TripLifecyclePage> createState() => _TripLifecyclePageState();
}

class _TripLifecyclePageState extends State<TripLifecyclePage> {
  RiderTripStatus _status = RiderTripStatus.searchingDriver;
  Timer? _pollTimer;
  bool _isPolling = false;
  int _finalFareCents = 0;
  String _currency = '';
  String _driverId = '';
  bool _trackingStarted = false;
  String? _pollError;
  bool _isPaying = false;

  @override
  void initState() {
    super.initState();
    _poll();
    _pollTimer = Timer.periodic(const Duration(seconds: 5), (_) => _poll());
  }

  @override
  void dispose() {
    _pollTimer?.cancel();
    super.dispose();
  }

  Future<void> _poll() async {
    if (_isPolling) return;
    _isPolling = true;
    try {
      final detail =
          await TripRepository(widget.apiClient).getTrip(widget.tripId);
      if (!mounted) return;
      final newStatus = _mapStatus(detail.status);
      setState(() {
        _status = newStatus;
        _finalFareCents = detail.finalFareCents;
        _currency = detail.currency;
        _driverId = detail.driverId;
        _pollError = null;
      });
      if (newStatus == RiderTripStatus.driverAssigned &&
          !_trackingStarted &&
          detail.driverId.isNotEmpty) {
        _trackingStarted = true;
        widget.onDriverAssigned?.call(detail.driverId);
      }
      if (newStatus.isTerminal) {
        _pollTimer?.cancel();
        _pollTimer = null;
      }
    } on ApiException catch (e) {
      if (mounted) setState(() => _pollError = e.message);
    } catch (_) {
      if (mounted) setState(() => _pollError = 'Network error. Retrying…');
    } finally {
      _isPolling = false;
    }
  }

  RiderTripStatus _mapStatus(String s) => switch (s) {
        'pending' || 'searching' => RiderTripStatus.searchingDriver,
        'driver_assigned' => RiderTripStatus.driverAssigned,
        'driver_arrived' => RiderTripStatus.driverArriving,
        'in_progress' => RiderTripStatus.inProgress,
        'completed' => RiderTripStatus.completed,
        'cancelled' => RiderTripStatus.cancelled,
        'payment_pending' => RiderTripStatus.paymentPending,
        'payment_success' => RiderTripStatus.paymentSuccess,
        'settled' => RiderTripStatus.settled,
        _ => _status,
      };

  Future<void> _pay(String method) async {
    if (_isPaying) return;
    setState(() => _isPaying = true);
    try {
      await TripRepository(widget.apiClient)
          .payRide(widget.tripId, paymentMethod: method);
    } on ApiException catch (e) {
      if (mounted) setState(() => _pollError = e.message);
    } catch (_) {
      if (mounted) setState(() => _pollError = 'Payment failed. Retrying…');
    } finally {
      if (mounted) setState(() => _isPaying = false);
    }
  }

  String get _fareText {
    if (_finalFareCents <= 0) return '—';
    final amount = _finalFareCents / 100.0;
    final sym = _currency.toUpperCase() == 'USD' ? r'$' : _currency;
    return '$sym${amount.toStringAsFixed(2)}';
  }

  void _cancelRide() {
    _pollTimer?.cancel();
    _pollTimer = null;
    // Best-effort: fire-and-forget so the UI isn't blocked.
    TripRepository(widget.apiClient).cancelRide(widget.tripId).ignore();
    Navigator.of(context).pop();
  }
  void _finish() => Navigator.of(context).pop();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Your Trip'),
        automaticallyImplyLeading: false,
      ),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  if (_pollError != null)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: Text(
                        _pollError!,
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.error,
                          fontSize: 13,
                        ),
                        textAlign: TextAlign.center,
                      ),
                    ),
                  AnimatedSwitcher(
                    duration: const Duration(milliseconds: 400),
                    transitionBuilder: (child, animation) => FadeTransition(
                      opacity: animation,
                      child: SlideTransition(
                        position: Tween<Offset>(
                          begin: const Offset(0, 0.05),
                          end: Offset.zero,
                        ).animate(animation),
                        child: child,
                      ),
                    ),
                    child: _buildCurrentView(),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildCurrentView() {
    final driver = MockTripCatalog.sampleDriver;
    return switch (_status) {
      RiderTripStatus.searchingDriver => SearchingDriverView(
          key: const ValueKey(RiderTripStatus.searchingDriver),
          tripSelection: widget.tripSelection,
          onCancel: _cancelRide,
        ),
      RiderTripStatus.driverAssigned => DriverAssignedView(
          key: const ValueKey(RiderTripStatus.driverAssigned),
          tripSelection: widget.tripSelection,
          driver: driver,
          onCancel: _cancelRide,
        ),
      RiderTripStatus.driverArriving => DriverArrivingView(
          key: const ValueKey(RiderTripStatus.driverArriving),
          tripSelection: widget.tripSelection,
          driver: driver,
          onCancel: _cancelRide,
        ),
      RiderTripStatus.inProgress => TripInProgressView(
          key: const ValueKey(RiderTripStatus.inProgress),
          tripSelection: widget.tripSelection,
          driver: driver,
        ),
      RiderTripStatus.completed => TripCompletedView(
          key: const ValueKey(RiderTripStatus.completed),
          tripSelection: widget.tripSelection,
          driver: driver,
          fareText: _fareText,
          onDone: _finish,
        ),
      RiderTripStatus.cancelled => TripCancelledView(
          key: const ValueKey(RiderTripStatus.cancelled),
          onDone: _finish,
        ),
      RiderTripStatus.paymentPending => _PaymentPendingView(
          key: const ValueKey(RiderTripStatus.paymentPending),
          fareText: _fareText,
          isPaying: _isPaying,
          onPayCash: () => _pay('cash'),
          onPayWallet: () => _pay('wallet'),
        ),
      RiderTripStatus.paymentSuccess ||
      RiderTripStatus.settled =>
        _PaymentSuccessView(
          key: const ValueKey('payment_done'),
          fareText: _fareText,
          onDone: _finish,
        ),
    };
  }
}

// ─── Payment views ────────────────────────────────────────────────────────────

class _PaymentPendingView extends StatelessWidget {
  const _PaymentPendingView({
    super.key,
    required this.fareText,
    required this.isPaying,
    required this.onPayCash,
    required this.onPayWallet,
  });

  final String fareText;
  final bool isPaying;
  final VoidCallback onPayCash;
  final VoidCallback onPayWallet;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.payment, size: 64, color: cs.primary),
          const SizedBox(height: 12),
          Text(
            'Payment',
            style: theme.textTheme.headlineSmall
                ?.copyWith(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          Text(
            'Please pay to complete your trip',
            style: theme.textTheme.bodyMedium
                ?.copyWith(color: cs.onSurfaceVariant),
          ),
          const SizedBox(height: 24),
          Card(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Total fare',
                      style: theme.textTheme.bodyLarge),
                  Text(
                    fareText,
                    style: theme.textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                      color: cs.primary,
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 24),
          if (isPaying)
            const CircularProgressIndicator()
          else ...[
            FilledButton.icon(
              onPressed: onPayCash,
              icon: const Icon(Icons.money),
              label: const Text('Pay with Cash'),
              style: FilledButton.styleFrom(
                minimumSize: const Size.fromHeight(52),
              ),
            ),
            const SizedBox(height: 12),
            OutlinedButton.icon(
              onPressed: onPayWallet,
              icon: const Icon(Icons.account_balance_wallet_outlined),
              label: const Text('Pay with Wallet'),
              style: OutlinedButton.styleFrom(
                minimumSize: const Size.fromHeight(52),
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _PaymentSuccessView extends StatelessWidget {
  const _PaymentSuccessView({
    super.key,
    required this.fareText,
    required this.onDone,
  });

  final String fareText;
  final VoidCallback onDone;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.verified, size: 72, color: cs.primary),
          const SizedBox(height: 12),
          Text(
            'Payment Complete',
            style: theme.textTheme.headlineSmall?.copyWith(
              fontWeight: FontWeight.bold,
              color: cs.primary,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            fareText,
            style: theme.textTheme.headlineMedium?.copyWith(
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Thank you for riding with FAIRRIDE!',
            style: theme.textTheme.bodyMedium
                ?.copyWith(color: cs.onSurfaceVariant),
          ),
          const SizedBox(height: 32),
          FilledButton(
            onPressed: onDone,
            style: FilledButton.styleFrom(
              minimumSize: const Size.fromHeight(52),
            ),
            child: const Text('Done'),
          ),
        ],
      ),
    );
  }
}
