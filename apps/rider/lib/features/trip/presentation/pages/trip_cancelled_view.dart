import 'package:flutter/material.dart';

/// Shown when the backend reports `cancelled` as the final trip state.
class TripCancelledView extends StatelessWidget {
  const TripCancelledView({super.key, required this.onDone});

  final VoidCallback onDone;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 40),
        Icon(Icons.cancel_outlined, size: 72, color: colorScheme.error),
        const SizedBox(height: 16),
        Text(
          'Chuyến đi đã hủy',
          style: Theme.of(context).textTheme.headlineSmall,
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 8),
        Text(
          'Chuyến đi của bạn đã bị hủy.',
          style: Theme.of(context)
              .textTheme
              .bodyMedium
              ?.copyWith(color: colorScheme.onSurfaceVariant),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: 40),
        FilledButton(onPressed: onDone, child: const Text('Xong')),
      ],
    );
  }
}
