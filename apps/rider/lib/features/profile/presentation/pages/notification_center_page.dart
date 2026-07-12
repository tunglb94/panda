import 'package:flutter/material.dart';

import 'package:rider/core/network/api_client.dart';

import '../../../../shared/widgets/app_empty_state.dart';
import '../../../../shared/widgets/async_state_view.dart';
import '../../data/notification_repository.dart';
import '../../domain/models/notification_item.dart';
import '../widgets/notification_tile.dart';

/// Notification Center: real in-app notifications from the Communication
/// Module (Part 3 — no more mock data). Loading / Success / Empty / Error
/// states via [AsyncStateView].
class NotificationCenterPage extends StatefulWidget {
  const NotificationCenterPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<NotificationCenterPage> createState() => _NotificationCenterPageState();
}

class _NotificationCenterPageState extends State<NotificationCenterPage> {
  late final NotificationRepository _repository = NotificationRepository(widget.apiClient);

  late Future<List<NotificationItem>> _future;
  List<NotificationItem>? _items;
  int _unreadCount = 0;

  @override
  void initState() {
    super.initState();
    _load();
  }

  void _load() {
    setState(() {
      _future = _repository.fetchAll().then((feed) {
        _items = feed.items;
        _unreadCount = feed.unreadCount;
        return feed.items;
      });
    });
  }

  void _markRead(NotificationItem item) {
    final items = _items;
    if (items == null || item.isRead) return;
    setState(() {
      final index = items.indexWhere((n) => n.id == item.id);
      if (index != -1) items[index] = items[index].copyWith(isRead: true);
      if (_unreadCount > 0) _unreadCount--;
    });
    // Best-effort — the local optimistic update above is what the UI
    // reflects immediately; a failed API call just means the item shows
    // read on this device but the server still has it unread (picked up
    // again on the next fetch), never a blocking error.
    _repository.markRead(item.id).catchError((_) {});
  }

  void _markAllRead() {
    final items = _items;
    if (items == null) return;
    final unread = items.where((n) => !n.isRead).toList();
    setState(() {
      for (var i = 0; i < items.length; i++) {
        items[i] = items[i].copyWith(isRead: true);
      }
      _unreadCount = 0;
    });
    for (final item in unread) {
      _repository.markRead(item.id).catchError((_) {});
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const BackButtonIcon(),
          tooltip: 'Quay lại',
          onPressed: () => Navigator.of(context).pop(_unreadCount),
        ),
        title: const Text('Thông báo'),
        actions: [
          IconButton(
            icon: const Icon(Icons.done_all),
            tooltip: 'Đánh dấu đã đọc tất cả',
            onPressed: _items == null || _unreadCount == 0 ? null : _markAllRead,
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
              emptyBuilder: (context) => const AppEmptyState(
                icon: Icons.notifications_none,
                title: 'Chưa có thông báo nào',
                subtitle: 'Chúng tôi sẽ báo cho bạn khi có thông báo mới.',
                mascotAsset: 'mascot_notification_empty.png',
              ),
              errorBuilder: (context, error) => AppEmptyState.error(
                title: 'Không thể tải thông báo',
                subtitle: error is ApiException && error.statusCode == 0 ? error.message : 'Vui lòng thử lại.',
                mascotAsset: 'mascot_no_connection.png',
                onAction: _load,
              ),
              successBuilder: (context, items) => RefreshIndicator(
                onRefresh: () async => _load(),
                child: ListView.separated(
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
      ),
    );
  }
}
