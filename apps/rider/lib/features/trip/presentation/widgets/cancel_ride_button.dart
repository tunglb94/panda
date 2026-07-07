import 'package:flutter/material.dart';

/// Cancel Ride action, shown only before the trip has actually started
/// (searching / assigned / arriving — see `RiderTripStatus.isCancellable`).
///
/// Confirms with the rider first; this never sends a cancellation request to
/// any backend — [onCancel] is a purely local, UI-level callback.
class CancelRideButton extends StatelessWidget {
  const CancelRideButton({super.key, required this.onCancel});

  final VoidCallback onCancel;

  @override
  Widget build(BuildContext context) {
    return OutlinedButton.icon(
      onPressed: () => _confirmCancel(context),
      icon: const Icon(Icons.close),
      label: const Text('Cancel Ride'),
      style: OutlinedButton.styleFrom(
        foregroundColor: Colors.red.shade600,
        side: BorderSide(color: Colors.red.shade200),
        minimumSize: const Size.fromHeight(52),
      ),
    );
  }

  Future<void> _confirmCancel(BuildContext context) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Cancel this ride?'),
        content: const Text(
          'This is a UI-only mock — no cancellation request is sent to the '
          'backend yet.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: const Text('Keep ride'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, true),
            child: const Text('Cancel ride'),
          ),
        ],
      ),
    );
    if (confirmed == true) onCancel();
  }
}
