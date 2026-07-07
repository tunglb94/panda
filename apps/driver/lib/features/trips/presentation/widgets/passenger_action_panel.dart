import 'package:flutter/material.dart';

/// Actions available while waiting at the pickup point: a primary
/// "Passenger On Board" button (mock only — `passengerBoarding` isn't a
/// state yet, see `TripOfferState`) plus Cancel Trip / Contact Rider, same
/// plain-callback-only contract as `DriverNavigationCard`'s actions (no
/// phone dialer, no popup/dialog).
class PassengerActionPanel extends StatelessWidget {
  const PassengerActionPanel({
    super.key,
    required this.onPassengerOnBoard,
    required this.onContactRider,
    required this.onCancelTrip,
  });

  final VoidCallback onPassengerOnBoard;
  final VoidCallback onContactRider;
  final VoidCallback onCancelTrip;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        FilledButton.icon(
          onPressed: onPassengerOnBoard,
          icon: const Icon(Icons.airline_seat_recline_normal),
          label: const Text('Passenger On Board'),
        ),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: OutlinedButton.icon(
                onPressed: onCancelTrip,
                icon: const Icon(Icons.close),
                label: const Text('Cancel Trip'),
                style: OutlinedButton.styleFrom(foregroundColor: Colors.red.shade600),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: OutlinedButton.icon(
                onPressed: onContactRider,
                icon: const Icon(Icons.call_outlined),
                label: const Text('Contact Rider'),
              ),
            ),
          ],
        ),
      ],
    );
  }
}
