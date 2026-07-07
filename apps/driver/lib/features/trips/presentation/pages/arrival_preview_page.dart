import 'package:flutter/material.dart';

import '../../domain/models/mock_trip_offer_catalog.dart';
import '../widgets/driver_arrival_card.dart';

/// Previews the Arrived-at-Pickup screen at fixed waiting durations
/// (Arrived / Waiting 00:00 / Waiting 03:00 / Waiting 08:00) without
/// touching the repository — each step just seeds `DriverArrivalCard` with
/// a different `initialWaitingSeconds`; its `WaitingTimer` still ticks
/// live from there, same as how other preview pages keep their embedded
/// widgets interactive rather than freezing them. Reached from
/// `TripOfferPreviewMenuPage`'s "Arrival Preview" entry.
class ArrivalPreviewPage extends StatefulWidget {
  const ArrivalPreviewPage({super.key});

  @override
  State<ArrivalPreviewPage> createState() => _ArrivalPreviewPageState();
}

enum _ArrivalPreviewStep { arrived, waiting0, waiting3, waiting8 }

extension on _ArrivalPreviewStep {
  String get label => switch (this) {
        _ArrivalPreviewStep.arrived => 'Arrived',
        _ArrivalPreviewStep.waiting0 => 'Waiting 00:00',
        _ArrivalPreviewStep.waiting3 => 'Waiting 03:00',
        _ArrivalPreviewStep.waiting8 => 'Waiting 08:00',
      };

  int get seconds => switch (this) {
        _ArrivalPreviewStep.arrived => 0,
        _ArrivalPreviewStep.waiting0 => 0,
        _ArrivalPreviewStep.waiting3 => 180,
        _ArrivalPreviewStep.waiting8 => 480,
      };
}

class _ArrivalPreviewPageState extends State<ArrivalPreviewPage> {
  _ArrivalPreviewStep _step = _ArrivalPreviewStep.arrived;

  @override
  Widget build(BuildContext context) {
    final offer = MockTripOfferCatalog.sample;

    return Scaffold(
      appBar: AppBar(title: const Text('Arrival Preview (dev)')),
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
                      for (final step in _ArrivalPreviewStep.values)
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
                        child: DriverArrivalCard(
                          offer: offer,
                          initialWaitingSeconds: _step.seconds,
                          onPassengerOnBoard: () {},
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
