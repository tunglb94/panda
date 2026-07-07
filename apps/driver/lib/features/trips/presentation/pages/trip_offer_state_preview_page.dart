import 'package:flutter/material.dart';

import '../../domain/models/mock_trip_offer_catalog.dart';
import '../../domain/models/trip_offer_state.dart';
import '../widgets/trip_offer_view.dart';

/// Renders a single [TripOfferState] in isolation, wrapped in its own
/// `Scaffold`, using the shared sample offer. This is the "independently
/// previewable" entry point requested for Phase D-03/D-04 — reachable from
/// `TripOfferPreviewMenuPage` (or `DispatchSessionPreviewPage`) without
/// needing to run the live countdown/accept flow on the Trips tab, and
/// without calling the repository.
class TripOfferStatePreviewPage extends StatelessWidget {
  const TripOfferStatePreviewPage({super.key, required this.state});

  final TripOfferState state;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('${state.label} (Preview)')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: TripOfferView(
                state: state,
                offer: MockTripOfferCatalog.sample,
                onAccept: () => Navigator.of(context).pop(),
                onReject: () => Navigator.of(context).pop(),
                onExpired: () {},
                onNavigate: () => ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Preview only — no real navigation.')),
                ),
                onRetry: () => Navigator.of(context).pop(),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
