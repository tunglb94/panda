import 'package:flutter/material.dart';

import '../../domain/models/trip_offer.dart';
import 'driver_status_banner.dart';
import 'fare_estimate_card.dart';
import 'passenger_action_panel.dart';
import 'route_stat_tile.dart';
import 'trip_address_row.dart';
import 'waiting_fee_card.dart';
import 'waiting_timer.dart';

/// The Arrived-at-Pickup screen (Phase D-06): "Arrived at Pickup" status
/// (via the existing, generic `DriverStatusBanner` — no new banner widget),
/// pickup address, passenger name, estimated fare (`FareEstimateCard`,
/// reused as-is), a live `WaitingTimer`, the mock `WaitingFeeCard` it feeds,
/// and `PassengerActionPanel`. [initialWaitingSeconds] lets the Arrival
/// Preview seed a fixed waiting duration without waiting in real time.
class DriverArrivalCard extends StatefulWidget {
  const DriverArrivalCard({
    super.key,
    required this.offer,
    required this.onPassengerOnBoard,
    required this.onContactRider,
    required this.onCancelTrip,
    this.initialWaitingSeconds = 0,
  });

  final TripOffer offer;
  final VoidCallback onPassengerOnBoard;
  final VoidCallback onContactRider;
  final VoidCallback onCancelTrip;
  final int initialWaitingSeconds;

  @override
  State<DriverArrivalCard> createState() => _DriverArrivalCardState();
}

class _DriverArrivalCardState extends State<DriverArrivalCard> {
  late int _elapsedMinutes;

  @override
  void initState() {
    super.initState();
    _elapsedMinutes = widget.initialWaitingSeconds ~/ 60;
  }

  @override
  Widget build(BuildContext context) {
    final offer = widget.offer;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          DriverStatusBanner(
            icon: Icons.flag_circle,
            title: 'Arrived at Pickup',
            subtitle: 'Waiting for ${offer.rider.name}',
          ),
          const SizedBox(height: 16),
          const Divider(height: 1),
          const SizedBox(height: 16),
          TripAddressRow(
            icon: Icons.circle,
            iconSize: 10,
            iconColor: Theme.of(context).colorScheme.primary,
            label: 'Pickup',
            address: offer.pickupAddress,
          ),
          const SizedBox(height: 12),
          RouteStatTile(
            icon: Icons.person_outline,
            label: 'Passenger',
            value: offer.rider.name,
          ),
          const SizedBox(height: 16),
          FareEstimateCard(offer: offer),
          const SizedBox(height: 16),
          WaitingTimer(
            initialSeconds: widget.initialWaitingSeconds,
            onMinutePassed: (minute) => setState(() => _elapsedMinutes = minute),
          ),
          const SizedBox(height: 12),
          WaitingFeeCard(elapsedMinutes: _elapsedMinutes),
          const SizedBox(height: 20),
          PassengerActionPanel(
            onPassengerOnBoard: widget.onPassengerOnBoard,
            onContactRider: widget.onContactRider,
            onCancelTrip: widget.onCancelTrip,
          ),
        ],
      ),
    );
  }
}
