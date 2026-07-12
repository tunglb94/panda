import 'package:flutter/material.dart';

import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/storage/token_storage.dart';
import 'package:rider/core/theme/app_spacing.dart';
import 'package:rider/features/history/presentation/pages/trip_history_page.dart';
import 'package:rider/features/wallet/presentation/pages/wallet_page.dart';
import 'package:rider/shared/widgets/app_badge.dart';
import 'package:rider/shared/widgets/app_card.dart';
import 'package:rider/shared/widgets/app_settings_tile.dart';
import 'package:rider/shared/widgets/async_state_view.dart';

import '../../data/notification_repository.dart';
import '../../domain/models/mock_profile_repository.dart';
import '../../domain/models/rider_profile.dart';
import '../widgets/profile_header.dart';
import '../widgets/profile_stats_row.dart';
import 'notification_center_page.dart';
import 'settings_page.dart';

/// Profile Screen: avatar, full name, phone number, member level, rating,
/// and total completed trips — fetched from [MockProfileRepository] so the
/// screen has a genuine Loading → Success transition. All data is mock; see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` Rider App Roadmap stage R8.
class ProfilePage extends StatefulWidget {
  const ProfilePage({
    super.key,
    required this.authState,
    required this.tokenStorage,
    required this.apiClient,
  });

  final AuthState authState;
  final TokenStorage tokenStorage;
  final ApiClient apiClient;

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  static const _profileRepository = MockProfileRepository();

  late Future<RiderProfile> _profileFuture;
  int _unreadCount = 0;

  @override
  void initState() {
    super.initState();
    _profileFuture = _profileRepository.fetchProfile();
    _loadUnreadCount();
  }

  Future<void> _loadUnreadCount() async {
    try {
      final feed = await NotificationRepository(widget.apiClient).fetchAll();
      if (mounted) setState(() => _unreadCount = feed.unreadCount);
    } catch (_) {
      // Non-fatal — badge just stays at 0 until the notification page loads.
    }
  }

  Future<void> _openNotifications() async {
    final result = await Navigator.of(context).push<int>(
      MaterialPageRoute(builder: (_) => NotificationCenterPage(apiClient: widget.apiClient)),
    );
    if (result != null && mounted) {
      setState(() => _unreadCount = result);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Hồ sơ'),
        actions: [
          IconButton(
            tooltip: 'Thông báo',
            onPressed: _openNotifications,
            icon: Stack(
              clipBehavior: Clip.none,
              children: [
                const Icon(Icons.notifications_outlined),
                Positioned(
                  right: -4,
                  top: -4,
                  child: AppBadge(count: _unreadCount),
                ),
              ],
            ),
          ),
        ],
      ),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(AppSpacing.lg),
              child: AsyncStateView<RiderProfile>(
                future: _profileFuture,
                successBuilder: (context, profile) => Column(
                  children: [
                    ProfileHeader(profile: profile),
                    const SizedBox(height: AppSpacing.xl),
                    ProfileStatsRow(profile: profile),
                    const SizedBox(height: AppSpacing.xxl),
                    AppCard(
                      padding: EdgeInsets.zero,
                      child: Column(
                        children: [
                          AppSettingsTile(
                            icon: Icons.account_balance_wallet_outlined,
                            label: 'Ví',
                            onTap: () => Navigator.of(context).push(
                              MaterialPageRoute(builder: (_) => const WalletPage()),
                            ),
                          ),
                          const Divider(height: 1),
                          AppSettingsTile(
                            icon: Icons.receipt_long_outlined,
                            label: 'Lịch sử chuyến đi',
                            onTap: () => Navigator.of(context).push(
                              MaterialPageRoute(
                                builder: (_) => TripHistoryPage(apiClient: widget.apiClient),
                              ),
                            ),
                          ),
                          const Divider(height: 1),
                          AppSettingsTile(
                            icon: Icons.settings_outlined,
                            label: 'Cài đặt',
                            onTap: () => Navigator.of(context).push(
                              MaterialPageRoute(builder: (_) => SettingsPage(apiClient: widget.apiClient)),
                            ),
                          ),
                          const Divider(height: 1),
                          AppSettingsTile(
                            icon: Icons.logout,
                            label: 'Đăng xuất',
                            isDestructive: true,
                            onTap: () async {
                              await widget.authState.logout(widget.tokenStorage);
                              // GoRouter's refreshListenable redirects to /login.
                            },
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
