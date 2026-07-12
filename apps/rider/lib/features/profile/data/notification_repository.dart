import 'package:rider/core/network/api_client.dart';

import '../domain/models/notification_item.dart';

/// Real in-app notification feed (Part 3 — no more mock/derived data).
/// Backed by the Communication Module's `GET /api/v1/notifications`.
class NotificationRepository {
  const NotificationRepository(this._client);

  final ApiClient _client;

  Future<NotificationFeed> fetchAll() async {
    final body = await _client.get('/api/v1/notifications');
    final raw = (body['notifications'] as List<dynamic>?) ?? const [];
    return NotificationFeed(
      items: raw.map((e) => NotificationItem.fromJson(e as Map<String, dynamic>)).toList(),
      unreadCount: (body['unread_count'] as num?)?.toInt() ?? 0,
    );
  }

  Future<void> markRead(String id) async {
    await _client.post('/api/v1/notifications/$id/read');
  }
}

class NotificationFeed {
  const NotificationFeed({required this.items, required this.unreadCount});

  final List<NotificationItem> items;
  final int unreadCount;
}
