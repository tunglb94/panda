import 'package:flutter/material.dart';

/// Status banner for the post-Accept dispatch session: shown for `failed`
/// and `timeout`. [onRetry] is required for both — the task's error
/// handling always ends in a Retry back to New Offer, never a silent
/// auto-revert.
class DispatchStatusBanner extends StatelessWidget {
  const DispatchStatusBanner({
    super.key,
    required this.icon,
    required this.color,
    required this.title,
    required this.message,
    required this.onRetry,
  });

  final IconData icon;
  final Color color;
  final String title;
  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 40, horizontal: 16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            padding: const EdgeInsets.all(18),
            decoration: BoxDecoration(color: color.withValues(alpha: 0.12), shape: BoxShape.circle),
            child: Icon(icon, size: 36, color: color),
          ),
          const SizedBox(height: 18),
          Text(title, style: const TextStyle(fontSize: 17, fontWeight: FontWeight.bold)),
          const SizedBox(height: 6),
          Text(
            message,
            textAlign: TextAlign.center,
            style: TextStyle(color: Colors.grey.shade600),
          ),
          const SizedBox(height: 20),
          SizedBox(
            width: double.infinity,
            child: OutlinedButton(onPressed: onRetry, child: const Text('Retry')),
          ),
        ],
      ),
    );
  }
}
