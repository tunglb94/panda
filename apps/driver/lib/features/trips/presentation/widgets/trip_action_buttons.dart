import 'package:flutter/material.dart';

/// Accept / Reject row for an incoming trip offer. Mock transitions only —
/// no backend call is made when either is tapped.
class TripActionButtons extends StatelessWidget {
  const TripActionButtons({
    super.key,
    required this.onAccept,
    required this.onReject,
  });

  final VoidCallback onAccept;
  final VoidCallback onReject;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: OutlinedButton(
            onPressed: onReject,
            style: OutlinedButton.styleFrom(
              foregroundColor: Colors.red.shade600,
              side: BorderSide(color: Colors.red.shade200),
            ),
            child: const Text('Reject'),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: FilledButton(
            onPressed: onAccept,
            child: const Text('Accept'),
          ),
        ),
      ],
    );
  }
}
