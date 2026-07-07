import 'package:flutter/material.dart';

import '../../domain/models/settings_entry.dart';
import 'settings_tile.dart';

/// A labelled group of [SettingsTile]s inside a card, e.g. "Account" or
/// "Preferences". Reused across every group on `SettingsPage`.
class SettingsSection extends StatelessWidget {
  const SettingsSection({
    super.key,
    this.title,
    required this.entries,
    required this.onTapEntry,
  });

  /// Section label, e.g. "Account". Omit for a standalone group (e.g. the
  /// Logout tile) that shouldn't have its own heading.
  final String? title;
  final List<SettingsEntry> entries;
  final ValueChanged<SettingsEntry> onTapEntry;

  @override
  Widget build(BuildContext context) {
    final label = title;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (label != null && label.isNotEmpty)
          Padding(
            padding: const EdgeInsets.only(bottom: 8, left: 4),
            child: Text(
              label,
              style: const TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: Color(0xFF6B7280),
              ),
            ),
          ),
        Container(
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: Colors.grey.shade200),
          ),
          padding: const EdgeInsets.symmetric(horizontal: 12),
          child: Column(
            children: [
              for (var i = 0; i < entries.length; i++) ...[
                if (i > 0) const Divider(height: 1),
                SettingsTile(
                  entry: entries[i],
                  onTap: () => onTapEntry(entries[i]),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }
}
