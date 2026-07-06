import 'package:flutter/material.dart';

enum NotificationType { trip, promotion, payment, system }

extension NotificationTypeX on NotificationType {
  IconData get icon => switch (this) {
        NotificationType.trip => Icons.directions_car,
        NotificationType.promotion => Icons.local_offer,
        NotificationType.payment => Icons.account_balance_wallet,
        NotificationType.system => Icons.info,
      };
}

/// A single Notification Center entry. Mock data only — there is no
/// Notification backend yet (see
/// `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.1).
class NotificationItem {
  const NotificationItem({
    required this.id,
    required this.type,
    required this.title,
    required this.body,
    required this.timestamp,
    required this.isRead,
  });

  final String id;
  final NotificationType type;
  final String title;
  final String body;
  final DateTime timestamp;
  final bool isRead;

  NotificationItem copyWith({bool? isRead}) => NotificationItem(
        id: id,
        type: type,
        title: title,
        body: body,
        timestamp: timestamp,
        isRead: isRead ?? this.isRead,
      );

  /// Coarse relative-time label (e.g. "12m ago"). Purely cosmetic — no
  /// locale-aware formatting package is introduced for this.
  String get relativeTimeLabel {
    final diff = DateTime.now().difference(timestamp);
    if (diff.inMinutes < 1) return 'Just now';
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    if (diff.inHours < 24) return '${diff.inHours}h ago';
    return '${diff.inDays}d ago';
  }
}
