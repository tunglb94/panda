import 'package:flutter/material.dart';

/// Placeholder for contacting the driver (call/message). Not implemented —
/// tapping it only shows a message; no HTTP request or backend dependency.
class ContactDriverButton extends StatelessWidget {
  const ContactDriverButton({super.key});

  @override
  Widget build(BuildContext context) {
    return FilledButton.tonalIcon(
      onPressed: () => _showPlaceholder(context),
      icon: const Icon(Icons.chat_bubble_outline),
      label: const Text('Contact Driver'),
      style: FilledButton.styleFrom(minimumSize: const Size.fromHeight(48)),
    );
  }

  void _showPlaceholder(BuildContext context) {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text(
          'Contacting the driver is a placeholder — not yet wired to the '
          'backend.',
        ),
      ),
    );
  }
}
