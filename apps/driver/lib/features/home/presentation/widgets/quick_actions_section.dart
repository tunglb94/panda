import 'package:flutter/material.dart';

import 'quick_action_card.dart';

/// Lays out the 4 required Quick Actions (Earnings, Trip History, Support,
/// Vehicle) in a responsive 2-column grid. Every tap is placeholder
/// navigation only — a message explaining the screen isn't wired up yet, no
/// actual navigation, no backend call.
class QuickActionsSection extends StatelessWidget {
  const QuickActionsSection({super.key});

  static const _actions = [
    (icon: Icons.payments_outlined, label: 'Earnings'),
    (icon: Icons.receipt_long_outlined, label: 'Trip History'),
    (icon: Icons.support_agent_outlined, label: 'Support'),
    (icon: Icons.directions_car_outlined, label: 'Vehicle'),
  ];

  @override
  Widget build(BuildContext context) {
    return GridView.count(
      crossAxisCount: 2,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      mainAxisSpacing: 10,
      crossAxisSpacing: 10,
      childAspectRatio: 1.6,
      children: [
        for (final action in _actions)
          QuickActionCard(
            icon: action.icon,
            label: action.label,
            onTap: () => _showPlaceholder(context, action.label),
          ),
      ],
    );
  }

  void _showPlaceholder(BuildContext context, String label) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('$label is a placeholder — not yet implemented.')),
    );
  }
}
