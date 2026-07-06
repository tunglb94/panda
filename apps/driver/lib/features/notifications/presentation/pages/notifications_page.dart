import 'package:flutter/material.dart';

import 'package:driver/shared/widgets/placeholder_tab_content.dart';

/// Notifications tab placeholder. Will eventually show driver-facing
/// notifications (dispatch offers, payout confirmations, announcements).
class NotificationsPage extends StatelessWidget {
  const NotificationsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Notifications')),
      body: const PlaceholderTabContent(
        icon: Icons.notifications_outlined,
        title: 'Notifications',
        subtitle: 'Driver notifications will appear here in a future phase.',
      ),
    );
  }
}
