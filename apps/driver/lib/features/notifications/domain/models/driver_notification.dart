import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';

/// Notification categories. [tripUpdate]/[delivery]/[chat]/[call] are the
/// categories with a real data source (the Communication Module's
/// `GET /api/v1/notifications` — see `NotificationRepository`); the rest
/// have no backend source and simply have no entries until one exists — an
/// honest empty category, not a fabricated one.
enum NotificationCategory {
  tripUpdate,
  delivery,
  chat,
  call,
  payment,
  system,
  bonus,
  promotion,
  update,
  warning,
  support,
}

extension NotificationCategoryX on NotificationCategory {
  String get label => switch (this) {
        NotificationCategory.tripUpdate => 'Chuyến xe',
        NotificationCategory.delivery => 'Giao hàng',
        NotificationCategory.chat => 'Tin nhắn',
        NotificationCategory.call => 'Cuộc gọi',
        NotificationCategory.payment => 'Thanh toán',
        NotificationCategory.system => 'Hệ thống',
        NotificationCategory.bonus => 'Thưởng',
        NotificationCategory.promotion => 'Khuyến mãi',
        NotificationCategory.update => 'Cập nhật',
        NotificationCategory.warning => 'Cảnh báo',
        NotificationCategory.support => 'Hỗ trợ',
      };

  IconData get icon => switch (this) {
        NotificationCategory.tripUpdate => Icons.directions_car,
        NotificationCategory.delivery => Icons.local_shipping_outlined,
        NotificationCategory.chat => Icons.chat_bubble_outline,
        NotificationCategory.call => Icons.call,
        NotificationCategory.payment => Icons.payments_outlined,
        NotificationCategory.system => Icons.info_outline,
        NotificationCategory.bonus => Icons.card_giftcard,
        NotificationCategory.promotion => Icons.local_offer_outlined,
        NotificationCategory.update => Icons.system_update_outlined,
        NotificationCategory.warning => Icons.warning_amber_rounded,
        NotificationCategory.support => Icons.support_agent_outlined,
      };

  Color get color => switch (this) {
        NotificationCategory.tripUpdate => AppColors.primary,
        NotificationCategory.delivery => AppColors.info,
        NotificationCategory.chat => AppColors.primary,
        NotificationCategory.call => AppColors.info,
        NotificationCategory.payment => AppColors.info,
        NotificationCategory.system => AppColors.textSecondary,
        NotificationCategory.bonus => AppColors.warning,
        NotificationCategory.promotion => const Color(0xFFEC4899),
        NotificationCategory.update => AppColors.info,
        NotificationCategory.warning => AppColors.error,
        NotificationCategory.support => AppColors.primary,
      };

  static NotificationCategory fromWire(String category) => switch (category) {
        'trip' => NotificationCategory.tripUpdate,
        'delivery' => NotificationCategory.delivery,
        'chat' => NotificationCategory.chat,
        'call' => NotificationCategory.call,
        'payment' => NotificationCategory.payment,
        'promotion' => NotificationCategory.promotion,
        _ => NotificationCategory.system,
      };
}

enum NotificationPriority { normal, high }

/// A single Notification Center entry — backed by the real Communication
/// Module API (`GET /api/v1/notifications`, see `NotificationRepository`).
class DriverNotification {
  const DriverNotification({
    required this.id,
    required this.category,
    required this.title,
    required this.subtitle,
    required this.timestamp,
    required this.isRead,
    this.priority = NotificationPriority.normal,
    this.tripId = '',
    this.conversationId = '',
  });

  final String id;
  final NotificationCategory category;
  final String title;
  final String subtitle;
  final DateTime timestamp;
  final bool isRead;
  final NotificationPriority priority;
  final String tripId;
  final String conversationId;

  DriverNotification copyWith({bool? isRead}) => DriverNotification(
        id: id,
        category: category,
        title: title,
        subtitle: subtitle,
        timestamp: timestamp,
        isRead: isRead ?? this.isRead,
        priority: priority,
        tripId: tripId,
        conversationId: conversationId,
      );

  factory DriverNotification.fromJson(Map<String, dynamic> json) => DriverNotification(
        id: json['id'] as String? ?? '',
        category: NotificationCategoryX.fromWire(json['category'] as String? ?? ''),
        title: json['title'] as String? ?? '',
        subtitle: json['body'] as String? ?? '',
        timestamp: DateTime.tryParse(json['created_at'] as String? ?? '')?.toLocal() ?? DateTime.now(),
        isRead: json['is_read'] as bool? ?? false,
        tripId: json['trip_id'] as String? ?? '',
        conversationId: json['conversation_id'] as String? ?? '',
      );
}
