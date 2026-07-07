import 'package:flutter/material.dart';

import '../../domain/models/mock_trip_offer_catalog.dart';
import '../../domain/models/route_progress_model.dart';
import '../widgets/driver_navigation_card.dart';
import '../widgets/trip_assigned_card.dart';

/// Previews the Assigned → Navigation flow at fixed progress steps
/// (100% / 80% / 60% / 40% / 20%) without touching the repository — every
/// value comes straight from `RouteProgressModel.mock`. Reached from
/// `TripOfferPreviewMenuPage`'s "Navigation Preview" entry.
class NavigationPreviewPage extends StatefulWidget {
  const NavigationPreviewPage({super.key});

  @override
  State<NavigationPreviewPage> createState() => _NavigationPreviewPageState();
}

enum _NavPreviewStep { assigned, p100, p80, p60, p40, p20 }

extension on _NavPreviewStep {
  String get label => switch (this) {
        _NavPreviewStep.assigned => 'Assigned',
        _NavPreviewStep.p100 => '100%',
        _NavPreviewStep.p80 => '80%',
        _NavPreviewStep.p60 => '60%',
        _NavPreviewStep.p40 => '40%',
        _NavPreviewStep.p20 => '20%',
      };

  int? get progress => switch (this) {
        _NavPreviewStep.assigned => null,
        _NavPreviewStep.p100 => 100,
        _NavPreviewStep.p80 => 80,
        _NavPreviewStep.p60 => 60,
        _NavPreviewStep.p40 => 40,
        _NavPreviewStep.p20 => 20,
      };
}

class _NavigationPreviewPageState extends State<NavigationPreviewPage> {
  _NavPreviewStep _step = _NavPreviewStep.assigned;

  @override
  Widget build(BuildContext context) {
    final offer = MockTripOfferCatalog.sample;
    final progress = _step.progress;

    return Scaffold(
      appBar: AppBar(title: const Text('Navigation Preview (dev)')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: Column(
              children: [
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                  child: Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      for (final step in _NavPreviewStep.values)
                        ChoiceChip(
                          label: Text(step.label),
                          selected: _step == step,
                          onSelected: (_) => setState(() => _step = step),
                        ),
                    ],
                  ),
                ),
                Expanded(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(16),
                    child: AnimatedSwitcher(
                      duration: const Duration(milliseconds: 350),
                      child: KeyedSubtree(
                        key: ValueKey(_step),
                        child: progress == null
                            ? TripAssignedCard(offer: offer, onNavigate: () {})
                            : DriverNavigationCard(
                                offer: offer,
                                route: RouteProgressModel.mock(
                                  progress: progress,
                                  trafficLevel: TrafficLevel.normal,
                                ),
                                onContactRider: () {},
                                onCancelTrip: () {},
                              ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
