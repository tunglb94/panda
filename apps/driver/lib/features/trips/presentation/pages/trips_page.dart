import 'package:flutter/material.dart';

import 'package:driver/shared/widgets/placeholder_tab_content.dart';

/// Trips tab placeholder. Will eventually show the active trip and past
/// trip history (Driver App Roadmap stages D5/D6).
class TripsPage extends StatelessWidget {
  const TripsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Trips')),
      body: const PlaceholderTabContent(
        icon: Icons.route_outlined,
        title: 'Trips',
        subtitle: 'Your active trip and trip history will appear here in a '
            'future phase.',
      ),
    );
  }
}
