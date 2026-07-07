import 'package:flutter/material.dart';

import '../../domain/models/trip_offer_state.dart';
import 'arrival_preview_page.dart';
import 'dispatch_session_preview_page.dart';
import 'navigation_preview_page.dart';
import 'trip_offer_state_preview_page.dart';

/// Lists the base offer-lifecycle states (New Offer / Rejected / Expired)
/// so each can be opened and previewed independently, plus links to the
/// "Dispatch Session Preview" (Accepting / Assigned / Failed / Timeout —
/// the states that only exist after Accept is pressed, Phase D-04), the
/// "Navigation Preview" (Assigned → Navigating at fixed progress steps,
/// Phase D-05), and the "Arrival Preview" (Arrived at fixed waiting
/// durations, Phase D-06).
class TripOfferPreviewMenuPage extends StatelessWidget {
  const TripOfferPreviewMenuPage({super.key});

  static const _offerStates = [
    TripOfferState.newOffer,
    TripOfferState.rejected,
    TripOfferState.expired,
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Trip Offer States (dev)')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          for (final state in _offerStates) ...[
            Card(
              child: ListTile(
                title: Text(state.label),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (_) => TripOfferStatePreviewPage(state: state),
                  ),
                ),
              ),
            ),
            const SizedBox(height: 8),
          ],
          const Divider(height: 24),
          Card(
            child: ListTile(
              leading: const Icon(Icons.sync_alt),
              title: const Text('Dispatch Session Preview'),
              subtitle: const Text('Accepting / Assigned / Failed / Timeout'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const DispatchSessionPreviewPage()),
              ),
            ),
          ),
          const SizedBox(height: 8),
          Card(
            child: ListTile(
              leading: const Icon(Icons.navigation_outlined),
              title: const Text('Navigation Preview'),
              subtitle: const Text('Assigned → Navigating (100% – 20%)'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const NavigationPreviewPage()),
              ),
            ),
          ),
          const SizedBox(height: 8),
          Card(
            child: ListTile(
              leading: const Icon(Icons.flag_circle_outlined),
              title: const Text('Arrival Preview'),
              subtitle: const Text('Arrived / Waiting 00:00 – 08:00'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => Navigator.of(context).push(
                MaterialPageRoute(builder: (_) => const ArrivalPreviewPage()),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
