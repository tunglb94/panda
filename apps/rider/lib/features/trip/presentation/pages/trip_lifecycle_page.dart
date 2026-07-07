import 'dart:async';

import 'package:flutter/material.dart';

import 'package:rider/features/booking/domain/models/mock_fare_calculator.dart';
import 'package:rider/features/map/domain/models/trip_selection.dart';

import '../../domain/models/mock_trip_catalog.dart';
import '../../domain/models/mock_trip_repository.dart';
import '../../domain/models/rider_trip_status.dart';
import 'driver_arriving_view.dart';
import 'driver_assigned_view.dart';
import 'searching_driver_view.dart';
import 'trip_completed_view.dart';
import 'trip_in_progress_view.dart';

/// Live trip lifecycle screen: one persistent Scaffold whose body animates
/// through the five [RiderTripStatus] states as [MockTripRepository] emits
/// them. Pushed after a mock booking request succeeds (see
/// `BookingFormBody._handleBookRide`) — this is the "reuse Booking UI" entry
/// point requested for Phase R-02.
///
/// No HTTP requests, no backend dependency: [repository] only emits mock
/// values on a timer.
class TripLifecyclePage extends StatefulWidget {
  const TripLifecyclePage({
    super.key,
    required this.tripSelection,
    required this.fare,
    this.repository = const MockTripRepository(),
  });

  final TripSelection tripSelection;
  final MockFareBreakdown fare;
  final MockTripRepository repository;

  @override
  State<TripLifecyclePage> createState() => _TripLifecyclePageState();
}

class _TripLifecyclePageState extends State<TripLifecyclePage> {
  RiderTripStatus _status = RiderTripStatus.searchingDriver;
  StreamSubscription<RiderTripStatus>? _subscription;

  @override
  void initState() {
    super.initState();
    _subscription = widget.repository.watchLifecycle().listen((status) {
      if (!mounted) return;
      setState(() => _status = status);
    });
  }

  @override
  void dispose() {
    _subscription?.cancel();
    super.dispose();
  }

  void _cancelRide() {
    // Mock only — no cancellation request is sent to any backend.
    Navigator.of(context).pop();
  }

  void _finish() {
    Navigator.of(context).pop();
  }

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
              child: AnimatedSwitcher(
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
          fare: widget.fare,
          onDone: _finish,
        ),
    };
  }
}
