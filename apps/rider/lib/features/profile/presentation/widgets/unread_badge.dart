import 'package:flutter/material.dart';

/// Small reusable unread-count badge. Renders nothing when [count] is 0.
/// Used on the Notifications bell icon (Profile app bar) and can be reused
/// anywhere else an unread count needs surfacing.
class UnreadBadge extends StatelessWidget {
  const UnreadBadge({super.key, required this.count});

  final int count;

  @override
  Widget build(BuildContext context) {
    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 200),
      transitionBuilder: (child, animation) =>
          ScaleTransition(scale: animation, child: child),
      child: count <= 0
          ? const SizedBox.shrink(key: ValueKey('no-badge'))
          : Container(
              key: ValueKey('badge-$count'),
              padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
              constraints: const BoxConstraints(minWidth: 16),
              decoration: BoxDecoration(
                color: Colors.red.shade600,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                count > 9 ? '9+' : '$count',
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 10,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
    );
  }
}
