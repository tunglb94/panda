import 'package:flutter/material.dart';

/// A compact "current activity" banner: icon in a tinted circle + title +
/// optional subtitle. Introduced for the Navigation screen's "Driving to
/// Pickup" status (Phase D-05), kept generic on purpose so later phases
/// (e.g. "Arrived", "Trip in Progress") can reuse it with a different
/// icon/title instead of a new one-off widget.
class DriverStatusBanner extends StatelessWidget {
  const DriverStatusBanner({super.key, required this.icon, required this.title, this.subtitle});

  final IconData icon;
  final String title;
  final String? subtitle;

  @override
  Widget build(BuildContext context) {
    final primary = Theme.of(context).colorScheme.primary;
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(10),
          decoration: BoxDecoration(color: primary.withValues(alpha: 0.12), shape: BoxShape.circle),
          child: Icon(icon, color: primary),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(title, style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
              if (subtitle != null) ...[
                const SizedBox(height: 2),
                Text(
                  subtitle!,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(fontSize: 12, color: Colors.grey.shade600),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }
}
