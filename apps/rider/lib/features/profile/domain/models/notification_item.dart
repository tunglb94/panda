import 'package:flutter/material.dart';

enum NotificationType { trip, delivery, chat, call, promotion, payment, system }

extension NotificationTypeX on NotificationType {
  IconData get icon => switch (this) {
        NotificationType.trip => Icons.directions_car,
        NotificationType.delivery => Icons.local_shipping,
        NotificationType.chat => Icons.chat_bubble,
        NotificationType.call => Icons.call,
        NotificationType.promotion => Icons.local_offer,
        NotificationType.payment => Icons.account_balance_wallet,
        NotificationType.system => Icons.info,
      };

  static NotificationType fromCategory(String category) => switch (category) {
        'trip' => NotificationType.trip,
        'delivery' => NotificationType.delivery,
        'chat' => NotificationType.chat,
        'call' => NotificationType.call,
        'promotion' => NotificationType.promotion,
        'payment' => NotificationType.payment,
        _ => NotificationType.system,
      };
}

/// A single Notification Center entry — backed by the real Communication
/// Module API (`GET /api/v1/notifications`, see `NotificationRepository`).
/// [tripId]/[conversationId] are optional deep-link targets: empty means
/// this notification has no associated trip/conversation to open.
class NotificationItem {
  const NotificationItem({
    required this.id,
    required this.type,
    required this.title,
    required this.body,
    required this.timestamp,
    required this.isRead,
    this.tripId = '',
    this.conversationId = '',
  });

  final String id;
  final NotificationType type;
  final String title;
  final String body;
  final DateTime timestamp;
  final bool isRead;
  final String tripId;
  final String conversationId;

  NotificationItem copyWith({bool? isRead}) => NotificationItem(
        id: id,
        type: type,
        title: title,
        body: body,
        timestamp: timestamp,
        isRead: isRead ?? this.isRead,
        tripId: tripId,
        conversationId: conversationId,
      );

  factory NotificationItem.fromJson(Map<String, dynamic> json) => NotificationItem(
        id: json['id'] as String? ?? '',
        type: NotificationTypeX.fromCategory(json['category'] as String? ?? ''),
        title: json['title'] as String? ?? '',
        body: json['body'] as String? ?? '',
        timestamp: DateTime.tryParse(json['created_at'] as String? ?? '')?.toLocal() ?? DateTime.now(),
        isRead: json['is_read'] as bool? ?? false,
        tripId: json['trip_id'] as String? ?? '',
        conversationId: json['conversation_id'] as String? ?? '',
      );

  /// Coarse relative-time label (e.g. "12m ago"). Purely cosmetic — no
  /// locale-aware formatting package is introduced for this.
  String get relativeTimeLabel {
    final diff = DateTime.now().difference(timestamp);
    if (diff.inMinutes < 1) return 'Vừa xong';
    if (diff.inMinutes < 60) return '${diff.inMinutes} phút trước';
    if (diff.inHours < 24) return '${diff.inHours} giờ trước';
    return '${diff.inDays} ngày trước';
  }
}
