import 'notification_item.dart';

/// Lets a caller preview all four required UI states for the Notification
/// Center without needing a real backend: [normal] returns the sample list,
/// [empty] returns no items, [error] throws. Selected from a dev-only menu
/// in `NotificationCenterPage`.
enum NotificationDemoMode { normal, empty, error }

/// Sample notifications returned by [MockNotificationRepository].
class MockNotificationCatalog {
  const MockNotificationCatalog._();

  /// Returns a fresh, independently-mutable list on every call (each
  /// [NotificationCenterPage] session gets its own copy to mark as read).
  static List<NotificationItem> sample() {
    final now = DateTime.now();
    return [
      NotificationItem(
        id: '1',
        type: NotificationType.trip,
        title: 'Trip completed',
        body: 'Your trip to Sample destination is complete. Thanks for riding with FAIRRIDE!',
        timestamp: now.subtract(const Duration(minutes: 12)),
        isRead: false,
      ),
      NotificationItem(
        id: '2',
        type: NotificationType.promotion,
        title: '20% off your next ride',
        body: 'Use code WELCOME20 before it expires.',
        timestamp: now.subtract(const Duration(hours: 3)),
        isRead: false,
      ),
      NotificationItem(
        id: '3',
        type: NotificationType.payment,
        title: 'Payment received',
        body: 'Your wallet was topped up (mock).',
        timestamp: now.subtract(const Duration(days: 1)),
        isRead: true,
      ),
      NotificationItem(
        id: '4',
        type: NotificationType.system,
        title: 'Welcome to FAIRRIDE',
        body: 'Thanks for joining. Explore the app to get started.',
        timestamp: now.subtract(const Duration(days: 2)),
        isRead: true,
      ),
    ];
  }
}

/// Mock repository for the Notification Center. No HTTP requests, no
/// backend dependency — see `docs/project/MVP_DEVELOPMENT_PLAN.md` §2.1.
class MockNotificationRepository {
  const MockNotificationRepository();

  Future<List<NotificationItem>> fetchNotifications({
    NotificationDemoMode mode = NotificationDemoMode.normal,
  }) async {
    await Future.delayed(const Duration(milliseconds: 800));
    switch (mode) {
      case NotificationDemoMode.error:
        throw StateError('Mock error: could not load notifications (simulated).');
      case NotificationDemoMode.empty:
        return const [];
      case NotificationDemoMode.normal:
        return MockNotificationCatalog.sample();
    }
  }
}
