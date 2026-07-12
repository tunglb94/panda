import 'package:flutter/material.dart';

import '../../../../core/theme/app_colors.dart';

/// The 8 notification categories requested for the Notification Center.
/// [tripUpdate] and [payment] are the only categories with a real data
/// source (derived from actual trip history — see
/// `NotificationRepository`); the rest have no backend source and simply
/// have no entries until one exists — an honest empty category, not a
/// fabricated one.
enum NotificationCategory {
  tripUpdate,
  delivery,
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
        NotificationCategory.payment => AppColors.info,
        NotificationCategory.system => AppColors.textSecondary,
        NotificationCategory.bonus => AppColors.warning,
        NotificationCategory.promotion => const Color(0xFFEC4899),
        NotificationCategory.update => AppColors.info,
        NotificationCategory.warning => AppColors.error,
        NotificationCategory.support => AppColors.primary,
      };
}

enum NotificationPriority { normal, high }

class DriverNotification {
  const DriverNotification({
    required this.id,
    required this.category,
    required this.title,
    required this.subtitle,
    required this.timestamp,
    required this.isRead,
    this.priority = NotificationPriority.normal,
  });

  final String id;
  final NotificationCategory category;
  final String title;
  final String subtitle;
  final DateTime timestamp;
  final bool isRead;
  final NotificationPriority priority;

  DriverNotification copyWith({bool? isRead}) => DriverNotification(
        id: id,
        category: category,
        title: title,
        subtitle: subtitle,
        timestamp: timestamp,
        isRead: isRead ?? this.isRead,
        priority: priority,
      );
}
