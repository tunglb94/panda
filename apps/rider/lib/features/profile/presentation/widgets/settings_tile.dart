import 'package:flutter/material.dart';

import '../../domain/models/settings_entry.dart';

/// Reusable settings row: icon + label + chevron, with an optional
/// destructive (red) style for actions like Logout.
class SettingsTile extends StatelessWidget {
  const SettingsTile({super.key, required this.entry, required this.onTap});

  final SettingsEntry entry;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final color = entry.isDestructive
        ? Colors.red.shade600
        : Theme.of(context).colorScheme.primary;
    return ListTile(
      contentPadding: EdgeInsets.zero,
      leading: Icon(entry.icon, color: color),
      title: Text(
        entry.label,
        style: entry.isDestructive ? TextStyle(color: color) : null,
      ),
      trailing: entry.isDestructive
          ? null
          : const Icon(Icons.chevron_right, color: Colors.grey),
      onTap: onTap,
    );
  }
}
