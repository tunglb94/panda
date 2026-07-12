import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';

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
  const SettingsPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Cài đặt')),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                SettingsSection(
                  title: 'Tài khoản',
                  entries: MockSettingsCatalog.account,
                  onTapEntry: (entry) => _handleEntry(context, entry),
                ),
                const SizedBox(height: 20),
                SettingsSection(
                  title: 'Tùy chọn',
                  entries: MockSettingsCatalog.preferences,
                  onTapEntry: (entry) => _handleEntry(context, entry),
                ),
                const SizedBox(height: 20),
                SettingsSection(
                  title: 'Hỗ trợ',
                  entries: MockSettingsCatalog.support,
                  onTapEntry: (entry) => _handleEntry(context, entry),
                ),
                const SizedBox(height: 20),
                DeveloperPreviewSection(apiClient: apiClient),
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
          MaterialPageRoute(builder: (_) => NotificationCenterPage(apiClient: apiClient)),
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
          SnackBar(content: Text('${entry.label} chưa được triển khai.')),
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
        title: const Text('Đăng xuất?'),
        content: const Text(
          'Đây chỉ là giao diện giả lập — chưa có phiên đăng nhập thực để '
          'đăng xuất (Identity chưa có endpoint đăng nhập).',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: const Text('Hủy'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, true),
            child: const Text('Đăng xuất'),
          ),
        ],
      ),
    );
    if (confirmed == true && context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Đã đăng xuất (giả lập — chưa có phiên backend thực).')),
      );
    }
  }
}
