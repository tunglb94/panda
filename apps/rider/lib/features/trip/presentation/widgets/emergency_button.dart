import 'package:flutter/material.dart';

/// Placeholder for an emergency/SOS action. Not implemented — tapping it
/// only shows a message explaining that, so the button's presence never
/// misleads the rider into thinking help was actually contacted.
class EmergencyButton extends StatelessWidget {
  const EmergencyButton({super.key});

  @override
  Widget build(BuildContext context) {
    return OutlinedButton.icon(
      onPressed: () => _showPlaceholder(context),
      icon: Icon(Icons.emergency_outlined, color: Colors.red.shade700),
      label: Text('Emergency', style: TextStyle(color: Colors.red.shade700)),
      style: OutlinedButton.styleFrom(
        side: BorderSide(color: Colors.red.shade200),
        minimumSize: const Size.fromHeight(48),
      ),
    );
  }

  void _showPlaceholder(BuildContext context) {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text(
          'Emergency assistance is a placeholder — not yet implemented.',
        ),
      ),
    );
  }
}
