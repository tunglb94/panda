import 'package:flutter/material.dart';

import 'package:rider/core/auth/auth_state.dart';
import 'package:rider/core/network/api_client.dart';
import 'package:rider/core/storage/token_storage.dart';
import 'package:rider/features/history/presentation/pages/trip_history_page.dart';

import '../../domain/models/mock_notification_repository.dart';
import '../../domain/models/mock_profile_repository.dart';
import '../../domain/models/rider_profile.dart';
import '../widgets/async_state_view.dart';
import '../widgets/profile_header.dart';
import '../widgets/profile_stats_row.dart';
import '../widgets/unread_badge.dart';
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
  int _unreadCount = MockNotificationCatalog.sample().where((n) => !n.isRead).length;

  @override
  void initState() {
    super.initState();
    _profileFuture = _profileRepository.fetchProfile();
  }

  Future<void> _openNotifications() async {
    final result = await Navigator.of(context).push<int>(
      MaterialPageRoute(builder: (_) => const NotificationCenterPage()),
    );
    if (result != null && mounted) {
      setState(() => _unreadCount = result);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Profile'),
        actions: [
          IconButton(
            tooltip: 'Notifications',
            onPressed: _openNotifications,
            icon: Stack(
              clipBehavior: Clip.none,
              children: [
                const Icon(Icons.notifications_outlined),
                Positioned(
                  right: -4,
                  top: -4,
                  child: UnreadBadge(count: _unreadCount),
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
              padding: const EdgeInsets.all(16),
              child: AsyncStateView<RiderProfile>(
                future: _profileFuture,
                successBuilder: (context, profile) => Column(
                  children: [
                    ProfileHeader(profile: profile),
                    const SizedBox(height: 20),
                    ProfileStatsRow(profile: profile),
                    const SizedBox(height: 24),
                    ListTile(
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                        side: BorderSide(color: Colors.grey.shade200),
                      ),
                      leading: Icon(
                        Icons.settings_outlined,
                        color: Theme.of(context).colorScheme.primary,
                      ),
                      title: const Text('Settings'),
                      trailing: const Icon(Icons.chevron_right, color: Colors.grey),
                      onTap: () => Navigator.of(context).push(
                        MaterialPageRoute(builder: (_) => const SettingsPage()),
                      ),
                    ),
                    const SizedBox(height: 12),
                    ListTile(
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                        side: BorderSide(color: Colors.grey.shade200),
                      ),
                      leading: Icon(
                        Icons.receipt_long_outlined,
                        color: Theme.of(context).colorScheme.primary,
                      ),
                      title: const Text('Trip History'),
                      trailing: const Icon(Icons.chevron_right, color: Colors.grey),
                      onTap: () => Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (_) => TripHistoryPage(apiClient: widget.apiClient),
                        ),
                      ),
                    ),
                    const SizedBox(height: 12),
                    ListTile(
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                        side: BorderSide(color: Colors.grey.shade200),
                      ),
                      leading: Icon(
                        Icons.logout,
                        color: Theme.of(context).colorScheme.error,
                      ),
                      title: Text(
                        'Sign Out',
                        style: TextStyle(
                            color: Theme.of(context).colorScheme.error),
                      ),
                      onTap: () async {
                        await widget.authState.logout(widget.tokenStorage);
                        // GoRouter's refreshListenable redirects to /login.
                      },
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
