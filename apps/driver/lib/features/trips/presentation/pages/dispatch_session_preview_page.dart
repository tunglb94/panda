import 'package:flutter/material.dart';

import '../../domain/models/trip_offer_state.dart';
import 'trip_offer_state_preview_page.dart';

/// Lists the post-Accept dispatch-session states (Accepting / Assigned /
/// Failed / Timeout) so each can be previewed independently, without
/// running the repository or the live 1.2s accept delay. Reached from
/// `TripOfferPreviewMenuPage`'s "Dispatch Session Preview" entry.
class DispatchSessionPreviewPage extends StatelessWidget {
  const DispatchSessionPreviewPage({super.key});

  static const _dispatchStates = [
    TripOfferState.accepting,
    TripOfferState.assigned,
    TripOfferState.failed,
    TripOfferState.timeout,
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Dispatch Session Preview (dev)')),
      body: ListView.separated(
        padding: const EdgeInsets.all(16),
        itemCount: _dispatchStates.length,
        separatorBuilder: (_, _) => const SizedBox(height: 8),
        itemBuilder: (context, index) {
          final state = _dispatchStates[index];
          return Card(
            child: ListTile(
              title: Text(state.label),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => TripOfferStatePreviewPage(state: state),
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}
