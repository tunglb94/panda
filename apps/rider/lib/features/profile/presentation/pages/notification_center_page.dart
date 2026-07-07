import 'package:flutter/material.dart';

import '../../domain/models/mock_notification_repository.dart';
import '../../domain/models/notification_item.dart';
import '../widgets/async_state_view.dart';
import '../widgets/notification_tile.dart';

/// Notification Center: fetches mock notifications and lets the rider mark
/// them read/unread locally. Demonstrates all four required UI states
/// (Loading / Success / Empty / Error) via [AsyncStateView] — Empty and
/// Error are reachable through the dev "Preview state" menu since the
/// static mock catalog is never naturally empty or failing.
class NotificationCenterPage extends StatefulWidget {
  const NotificationCenterPage({super.key});

  @override
  State<NotificationCenterPage> createState() => _NotificationCenterPageState();
}

class _NotificationCenterPageState extends State<NotificationCenterPage> {
  static const _repository = MockNotificationRepository();

  NotificationDemoMode _mode = NotificationDemoMode.normal;
  late Future<List<NotificationItem>> _future;
  List<NotificationItem>? _items;

  @override
  void initState() {
    super.initState();
    _load();
  }

  void _load() {
    setState(() {
      _future = _repository.fetchNotifications(mode: _mode).then((items) {
        _items = items;
        return items;
      });
    });
  }

  int get _unreadCount => _items?.where((n) => !n.isRead).length ?? 0;

  void _markRead(NotificationItem item) {
    final items = _items;
    if (items == null) return;
    setState(() {
      final index = items.indexWhere((n) => n.id == item.id);
      if (index != -1) items[index] = items[index].copyWith(isRead: true);
    });
  }

  void _markAllRead() {
    final items = _items;
    if (items == null) return;
    setState(() {
      for (var i = 0; i < items.length; i++) {
        items[i] = items[i].copyWith(isRead: true);
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const BackButtonIcon(),
          tooltip: 'Back',
          onPressed: () => Navigator.of(context).pop(_unreadCount),
        ),
        title: const Text('Notifications'),
        actions: [
          IconButton(
            icon: const Icon(Icons.done_all),
            tooltip: 'Mark all read',
            onPressed: _items == null || _unreadCount == 0 ? null : _markAllRead,
          ),
          PopupMenuButton<NotificationDemoMode>(
            tooltip: 'Preview state (dev)',
            icon: const Icon(Icons.tune),
            onSelected: (mode) {
              _mode = mode;
              _load();
            },
            itemBuilder: (context) => const [
              PopupMenuItem(value: NotificationDemoMode.normal, child: Text('Normal')),
              PopupMenuItem(value: NotificationDemoMode.empty, child: Text('Empty (dev)')),
              PopupMenuItem(value: NotificationDemoMode.error, child: Text('Error (dev)')),
            ],
          ),
        ],
      ),
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: AsyncStateView<List<NotificationItem>>(
              future: _future,
              isEmpty: (items) => items.isEmpty,
              emptyBuilder: (context) => Padding(
                padding: const EdgeInsets.symmetric(vertical: 48, horizontal: 24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.notifications_none, size: 48, color: Colors.grey.shade400),
                    const SizedBox(height: 12),
                    const Text('No notifications yet',
                        style: TextStyle(fontWeight: FontWeight.w600)),
                    const SizedBox(height: 4),
                    Text(
                      "We'll let you know when something new arrives.",
                      textAlign: TextAlign.center,
                      style: TextStyle(color: Colors.grey.shade500),
                    ),
                  ],
                ),
              ),
              errorBuilder: (context, error) => Padding(
                padding: const EdgeInsets.symmetric(vertical: 48, horizontal: 24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.error_outline, size: 48, color: Colors.red.shade400),
                    const SizedBox(height: 12),
                    const Text("Couldn't load notifications",
                        style: TextStyle(fontWeight: FontWeight.w600)),
                    const SizedBox(height: 12),
                    OutlinedButton(onPressed: _load, child: const Text('Retry')),
                  ],
                ),
              ),
              successBuilder: (context, items) => ListView.separated(
                padding: const EdgeInsets.all(16),
                itemCount: items.length,
                separatorBuilder: (_, _) => const SizedBox(height: 8),
                itemBuilder: (context, index) {
                  final item = items[index];
                  return NotificationTile(
                    item: item,
                    onTap: () => _markRead(item),
                  );
                },
              ),
            ),
          ),
        ),
      ),
    );
  }
}
