import 'package:flutter/material.dart';

import '../../domain/models/settings_entry.dart';
import '../widgets/developer_preview_section.dart';
import '../widgets/settings_section.dart';
import 'notification_center_page.dart';

/// Full settings list. Every entry except Notifications and Logout is a
/// placeholder in this phase (see `_handleEntry`) — none of Personal
/// Information / Payment Methods / Privacy / Security / Language / Help
/// Center / About have real screens yet; building those out is outside the
/// scope of this phase.
class SettingsPage extends StatelessWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Settings')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                SettingsSection(
                  title: 'Account',
                  entries: MockSettingsCatalog.account,
                  onTapEntry: (entry) => _handleEntry(context, entry),
                ),
                const SizedBox(height: 20),
                SettingsSection(
                  title: 'Preferences',
                  entries: MockSettingsCatalog.preferences,
                  onTapEntry: (entry) => _handleEntry(context, entry),
                ),
                const SizedBox(height: 20),
                SettingsSection(
                  title: 'Support',
                  entries: MockSettingsCatalog.support,
                  onTapEntry: (entry) => _handleEntry(context, entry),
                ),
                const SizedBox(height: 20),
                const DeveloperPreviewSection(),
                const SizedBox(height: 20),
                SettingsSection(
                  entries: [MockSettingsCatalog.logout],
                  onTapEntry: (entry) => _handleEntry(context, entry),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _handleEntry(BuildContext context, SettingsEntry entry) async {
    switch (entry.action) {
      case SettingsAction.notifications:
        await Navigator.of(context).push(
          MaterialPageRoute(builder: (_) => const NotificationCenterPage()),
        );
      case SettingsAction.logout:
        await _confirmLogout(context);
      case SettingsAction.personalInformation:
      case SettingsAction.paymentMethods:
      case SettingsAction.privacy:
      case SettingsAction.security:
      case SettingsAction.language:
      case SettingsAction.helpCenter:
      case SettingsAction.about:
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('${entry.label} is a placeholder — not yet implemented.')),
        );
      case SettingsAction.developerPreview:
        // Not routed through here — DeveloperPreviewSection handles its own
        // navigation directly. Kept for switch exhaustiveness.
        break;
    }
  }

  Future<void> _confirmLogout(BuildContext context) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Log out?'),
        content: const Text(
          'This is a UI-only mock — there is no real session to sign out of '
          'yet (Identity has no login endpoint).',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, true),
            child: const Text('Log out'),
          ),
        ],
      ),
    );
    if (confirmed == true && context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Logged out (mock — no backend session existed).')),
      );
    }
  }
}
