import 'package:flutter/material.dart';

import '../../domain/models/rider_trip_status.dart';
import 'trip_state_preview_page.dart';

/// Lists all five trip lifecycle states so each can be opened and previewed
/// independently during development, without waiting for the mock
/// lifecycle timer in `TripLifecyclePage` to reach it.
class TripPreviewMenuPage extends StatelessWidget {
  const TripPreviewMenuPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Trip UI Preview (dev)')),
      body: ListView.separated(
        padding: const EdgeInsets.all(16),
        itemCount: RiderTripStatus.values.length,
        separatorBuilder: (_, _) => const SizedBox(height: 8),
        itemBuilder: (context, index) {
          final status = RiderTripStatus.values[index];
          return Card(
            child: ListTile(
              leading: Icon(status.icon),
              title: Text(status.label),
              subtitle: Text(status.statusMessage),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) => TripStatePreviewPage(status: status),
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}
