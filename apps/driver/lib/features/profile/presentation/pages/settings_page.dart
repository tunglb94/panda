import 'package:flutter/material.dart';

import '../../../../core/auth/auth_state.dart';
import '../../../../core/storage/token_storage.dart';
import '../../../../core/theme/app_spacing.dart';
import '../../../../shared/widgets/app_card.dart';
import '../../../../shared/widgets/app_dialog.dart';
import '../../../../shared/widgets/app_settings_tile.dart';
import '../../../../shared/widgets/app_snackbar.dart';
import 'developer_page.dart';
import 'documents_page.dart';

/// Settings — every row either does something real (navigates to a real
/// screen, or the real logout) or is an honest, clearly-labeled
/// placeholder. Nothing here silently does nothing without telling the
/// driver why.
class SettingsPage extends StatelessWidget {
  const SettingsPage({
    super.key,
    required this.authState,
    required this.tokenStorage,
    required this.driverId,
    required this.buildVehiclePage,
  });

  final AuthState authState;
  final TokenStorage tokenStorage;
  final String driverId;

  /// Factory for `VehicleCenterPage` — passed in rather than constructed
  /// here so this page doesn't need its own `ApiClient` reference beyond
  /// what the caller already has.
  final Widget Function() buildVehiclePage;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Cài đặt')),
      body: ListView(
        padding: const EdgeInsets.all(AppSpacing.lg),
        children: [
          _SectionLabel('Tài khoản'),
          AppCard(
            padding: EdgeInsets.zero,
            child: Column(
              children: [
                AppSettingsTile(
                  icon: Icons.person_outline,
                  label: 'Thông tin cá nhân',
                  onTap: () => _placeholder(context, 'Chỉnh sửa thông tin cá nhân'),
                ),
                const Divider(height: 1),
                AppSettingsTile(
                  icon: Icons.directions_car_outlined,
                  label: 'Xe',
                  onTap: () => Navigator.of(context).push(
                    MaterialPageRoute(builder: (_) => buildVehiclePage()),
                  ),
                ),
                const Divider(height: 1),
                AppSettingsTile(
                  icon: Icons.account_balance_outlined,
                  label: 'Ngân hàng',
                  onTap: () => _placeholder(context, 'Liên kết ngân hàng'),
                ),
                const Divider(height: 1),
                AppSettingsTile(
                  icon: Icons.folder_outlined,
                  label: 'Giấy tờ (GPLX, bảo hiểm, kiểm định…)',
                  onTap: () => Navigator.of(context).push(
                    MaterialPageRoute(builder: (_) => const DocumentsPage()),
                  ),
                ),
                const Divider(height: 1),
                AppSettingsTile(
                  icon: Icons.lock_outline,
                  label: 'Đổi mật khẩu',
                  onTap: () => _explainPasswordless(context),
                ),
              ],
            ),
          ),
          const SizedBox(height: AppSpacing.xl),
          _SectionLabel('Tùy chọn'),
          AppCard(
            padding: EdgeInsets.zero,
            child: Column(
              children: [
                AppSettingsTile(
                  icon: Icons.notifications_outlined,
                  label: 'Thông báo',
                  onTap: () => _placeholder(context, 'Tùy chọn thông báo'),
                ),
                const Divider(height: 1),
                AppSettingsTile(
                  icon: Icons.language_outlined,
                  label: 'Ngôn ngữ',
                  trailing: const _ValueLabel('Tiếng Việt'),
                  onTap: () => AppDialog.info(
                    context,
                    title: 'Ngôn ngữ',
                    message: 'Ứng dụng hiện chỉ hỗ trợ Tiếng Việt.',
                  ),
                ),
                const Divider(height: 1),
                AppSettingsTile(
                  icon: Icons.palette_outlined,
                  label: 'Giao diện',
                  trailing: const _ValueLabel('Sáng'),
                  onTap: () => AppDialog.info(
                    context,
                    title: 'Giao diện',
                    message: 'Chế độ Tối chưa được hỗ trợ — sẽ ra mắt trong giai đoạn tiếp theo.',
                  ),
                ),
                const Divider(height: 1),
                AppSettingsTile(
                  icon: Icons.privacy_tip_outlined,
                  label: 'Quyền riêng tư',
                  onTap: () => _placeholder(context, 'Chính sách quyền riêng tư'),
                ),
                const Divider(height: 1),
                AppSettingsTile(
                  icon: Icons.description_outlined,
                  label: 'Điều khoản sử dụng',
                  onTap: () => _placeholder(context, 'Điều khoản sử dụng'),
                ),
              ],
            ),
          ),
          const SizedBox(height: AppSpacing.xl),
          _SectionLabel('Khác'),
          AppCard(
            padding: EdgeInsets.zero,
            child: Column(
              children: [
                AppSettingsTile(
                  icon: Icons.info_outline,
                  label: 'Phiên bản ứng dụng',
                  onTap: () => Navigator.of(context).push(
                    MaterialPageRoute(builder: (_) => const DeveloperPage()),
                  ),
                ),
                const Divider(height: 1),
                AppSettingsTile(
                  icon: Icons.logout,
                  label: 'Đăng xuất',
                  isDestructive: true,
                  onTap: () => _confirmLogout(context),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  void _placeholder(BuildContext context, String feature) {
    AppSnackbar.show(context, '$feature chưa khả dụng — sẽ ra mắt trong giai đoạn tiếp theo.');
  }

  void _explainPasswordless(BuildContext context) {
    AppDialog.info(
      context,
      title: 'Đổi mật khẩu',
      message: 'PandaDriver đăng nhập bằng số điện thoại, không dùng mật khẩu — '
          'nên không có bước đổi mật khẩu nào cần thiết.',
    );
  }

  Future<void> _confirmLogout(BuildContext context) async {
    final confirmed = await AppDialog.confirm(
      context,
      title: 'Đăng xuất khỏi PandaDriver?',
      message: 'Bạn sẽ cần đăng nhập lại bằng số điện thoại để tiếp tục nhận chuyến.',
      confirmLabel: 'Đăng xuất',
      isDestructive: true,
    );
    if (confirmed) {
      await authState.logout(tokenStorage);
      // GoRouter's refreshListenable redirects to /login automatically.
    }
  }
}

class _SectionLabel extends StatelessWidget {
  const _SectionLabel(this.label);

  final String label;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: AppSpacing.sm, left: 4),
      child: Text(label, style: Theme.of(context).textTheme.labelMedium),
    );
  }
}

class _ValueLabel extends StatelessWidget {
  const _ValueLabel(this.value);

  final String value;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(value, style: Theme.of(context).textTheme.bodySmall),
        const Icon(Icons.chevron_right, size: 18),
      ],
    );
  }
}
