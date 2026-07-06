import 'package:flutter/material.dart';

/// Identifies which settings row was tapped. Every action beyond
/// [notifications] and [logout] is a placeholder in this phase — see
/// `SettingsPage` for how each is handled.
enum SettingsAction {
  personalInformation,
  paymentMethods,
  notifications,
  privacy,
  security,
  language,
  helpCenter,
  about,
  logout,

  /// Not part of the required settings list — routes to the Phase R-02
  /// "Trip UI Preview (dev)" menu, kept in its own Developer section.
  developerPreview,
}

/// A single settings row: icon + label + which [SettingsAction] it triggers.
class SettingsEntry {
  const SettingsEntry({
    required this.action,
    required this.icon,
    required this.label,
    this.isDestructive = false,
  });

  final SettingsAction action;
  final IconData icon;
  final String label;

  /// Styles the tile as a destructive action (e.g. Logout) — red icon/text.
  final bool isDestructive;
}

/// Central catalog of settings entries, grouped for display in
/// `SettingsPage`.
class MockSettingsCatalog {
  const MockSettingsCatalog._();

  static const List<SettingsEntry> account = [
    SettingsEntry(
      action: SettingsAction.personalInformation,
      icon: Icons.person_outline,
      label: 'Personal Information',
    ),
    SettingsEntry(
      action: SettingsAction.paymentMethods,
      icon: Icons.payment_outlined,
      label: 'Payment Methods',
    ),
    SettingsEntry(
      action: SettingsAction.notifications,
      icon: Icons.notifications_outlined,
      label: 'Notifications',
    ),
  ];

  static const List<SettingsEntry> preferences = [
    SettingsEntry(
      action: SettingsAction.privacy,
      icon: Icons.privacy_tip_outlined,
      label: 'Privacy',
    ),
    SettingsEntry(
      action: SettingsAction.security,
      icon: Icons.lock_outline,
      label: 'Security',
    ),
    SettingsEntry(
      action: SettingsAction.language,
      icon: Icons.language_outlined,
      label: 'Language',
    ),
  ];

  static const List<SettingsEntry> support = [
    SettingsEntry(
      action: SettingsAction.helpCenter,
      icon: Icons.help_outline,
      label: 'Help Center',
    ),
    SettingsEntry(
      action: SettingsAction.about,
      icon: Icons.info_outline,
      label: 'About',
    ),
  ];

  static const SettingsEntry logout = SettingsEntry(
    action: SettingsAction.logout,
    icon: Icons.logout,
    label: 'Logout',
    isDestructive: true,
  );
}
