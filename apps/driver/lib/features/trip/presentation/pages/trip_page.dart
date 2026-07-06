import 'package:flutter/material.dart';

class TripPage extends StatelessWidget {
  const TripPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Trip')),
      body: const SafeArea(
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Column(
            children: [
              _ActiveTripCard(),
              SizedBox(height: 16),
              Expanded(child: _RecentTripsCard()),
            ],
          ),
        ),
      ),
    );
  }
}

class _ActiveTripCard extends StatelessWidget {
  const _ActiveTripCard();

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Active Trip', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 16),
            Center(
              child: Column(
                children: [
                  Icon(Icons.directions_car, size: 48, color: cs.outline),
                  const SizedBox(height: 8),
                  Text(
                    'No active trip',
                    style: Theme.of(context)
                        .textTheme
                        .bodyMedium
                        ?.copyWith(color: cs.onSurfaceVariant),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _RecentTripsCard extends StatelessWidget {
  const _RecentTripsCard();

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Recent Trips',
                style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 16),
            Center(
              child: Text(
                'No recent trips',
                style: Theme.of(context)
                    .textTheme
                    .bodyMedium
                    ?.copyWith(color: cs.onSurfaceVariant),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
