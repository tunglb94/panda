import 'package:flutter/material.dart';

import 'package:driver/core/network/api_client.dart';
import 'package:driver/core/theme/app_colors.dart';
import 'package:driver/core/theme/app_radius.dart';
import 'package:driver/core/theme/app_spacing.dart';
import 'package:driver/shared/widgets/app_empty_state.dart';
import 'package:driver/shared/widgets/app_skeleton.dart';

import '../../data/notification_repository.dart';
import '../../domain/models/driver_notification.dart';
import '../widgets/notification_tile.dart';

enum _QuickFilter { all, unread, tripUpdate, payment, promotion, system }

/// Notification Center — grouped by recency (Hôm nay/Hôm qua/7 ngày/Cũ
/// hơn), filterable, searchable. List content is real (derived from actual
/// trip history — see `NotificationRepository`); categories with no
/// backend source (Hệ thống/Thưởng/Khuyến mãi/Cập nhật/Cảnh báo/Hỗ trợ)
/// simply have zero entries rather than fabricated ones — filtering to one
/// of those categories today will honestly show "Không có thông báo".
///
/// Read/unread state is tracked in-memory for this session only (there is
/// no notification-read-state backend to persist it), matching the exact
/// pattern `apps/rider`'s NotificationCenterPage already uses for the same
/// reason.
class NotificationsPage extends StatefulWidget {
  const NotificationsPage({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<NotificationsPage> createState() => _NotificationsPageState();
}

class _NotificationsPageState extends State<NotificationsPage> {
  late final NotificationRepository _repo;
  late Future<List<DriverNotification>> _future;
  List<DriverNotification>? _items;
  _QuickFilter _filter = _QuickFilter.all;
  final _searchController = TextEditingController();
  String _query = '';

  @override
  void initState() {
    super.initState();
    _repo = NotificationRepository(widget.apiClient);
    _load();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _load() {
    setState(() {
      _future = _repo.fetchAll().then((items) {
        _items = items;
        return items;
      });
    });
  }

  int get _unreadCount => _items?.where((n) => !n.isRead).length ?? 0;

  void _markRead(DriverNotification n) {
    final items = _items;
    if (items == null) return;
    setState(() {
      final i = items.indexWhere((x) => x.id == n.id);
      if (i != -1) items[i] = items[i].copyWith(isRead: true);
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

  List<DriverNotification> _applyFilters(List<DriverNotification> items) {
    var result = items.where((n) {
      return switch (_filter) {
        _QuickFilter.all => true,
        _QuickFilter.unread => !n.isRead,
        _QuickFilter.tripUpdate => n.category == NotificationCategory.tripUpdate,
        _QuickFilter.payment => n.category == NotificationCategory.payment,
        _QuickFilter.promotion => n.category == NotificationCategory.promotion,
        _QuickFilter.system => n.category == NotificationCategory.system,
      };
    }).toList();

    if (_query.isNotEmpty) {
      final q = _query.toLowerCase();
      result = result
          .where((n) => n.title.toLowerCase().contains(q) || n.subtitle.toLowerCase().contains(q))
          .toList();
    }
    return result;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Thông báo'),
        actions: [
          IconButton(
            icon: const Icon(Icons.done_all),
            tooltip: 'Đánh dấu đã đọc tất cả',
            onPressed: (_items == null || _unreadCount == 0) ? null : _markAllRead,
          ),
        ],
      ),
      body: FutureBuilder<List<DriverNotification>>(
        future: _future,
        builder: (context, snap) {
          if (snap.connectionState == ConnectionState.waiting) {
            return ListView.separated(
              padding: const EdgeInsets.all(AppSpacing.lg),
              itemCount: 6,
              separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
              itemBuilder: (context, i) => const AppSkeletonListTile(),
            );
          }
          if (snap.hasError) {
            return AppEmptyState.error(
              subtitle: snap.error is ApiException && (snap.error as ApiException).statusCode == 0
                  ? (snap.error as ApiException).message
                  : 'Không thể tải thông báo.',
              onAction: _load,
            );
          }

          final all = snap.data ?? [];
          final filtered = _applyFilters(all);

          return Column(
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.md, AppSpacing.lg, 0),
                child: TextField(
                  controller: _searchController,
                  onChanged: (v) => setState(() => _query = v),
                  decoration: InputDecoration(
                    hintText: 'Tìm thông báo…',
                    prefixIcon: const Icon(Icons.search, size: 20),
                    suffixIcon: _query.isEmpty
                        ? null
                        : IconButton(
                            icon: const Icon(Icons.clear, size: 18),
                            tooltip: 'Xóa tìm kiếm',
                            onPressed: () => setState(() {
                              _searchController.clear();
                              _query = '';
                            }),
                          ),
                  ),
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(AppSpacing.lg),
                child: SizedBox(
                  height: 34,
                  child: ListView(
                    scrollDirection: Axis.horizontal,
                    children: [
                      _FilterChip(
                        label: 'Tất cả',
                        selected: _filter == _QuickFilter.all,
                        onTap: () => setState(() => _filter = _QuickFilter.all),
                      ),
                      _FilterChip(
                        label: 'Chưa đọc',
                        selected: _filter == _QuickFilter.unread,
                        onTap: () => setState(() => _filter = _QuickFilter.unread),
                      ),
                      _FilterChip(
                        label: 'Chuyến xe',
                        selected: _filter == _QuickFilter.tripUpdate,
                        onTap: () => setState(() => _filter = _QuickFilter.tripUpdate),
                      ),
                      _FilterChip(
                        label: 'Thanh toán',
                        selected: _filter == _QuickFilter.payment,
                        onTap: () => setState(() => _filter = _QuickFilter.payment),
                      ),
                      _FilterChip(
                        label: 'Khuyến mãi',
                        selected: _filter == _QuickFilter.promotion,
                        onTap: () => setState(() => _filter = _QuickFilter.promotion),
                      ),
                      _FilterChip(
                        label: 'Hệ thống',
                        selected: _filter == _QuickFilter.system,
                        onTap: () => setState(() => _filter = _QuickFilter.system),
                      ),
                    ],
                  ),
                ),
              ),
              Expanded(
                child: filtered.isEmpty
                    ? AppEmptyState(
                        icon: _query.isNotEmpty ? Icons.search_off : Icons.notifications_none,
                        title: _query.isNotEmpty ? 'Không tìm thấy kết quả' : 'Không có thông báo',
                        subtitle: _query.isNotEmpty
                            ? 'Thử từ khóa khác.'
                            : (all.isEmpty
                                ? 'Thông báo về chuyến đi và thanh toán sẽ xuất hiện ở đây.'
                                : 'Không có thông báo nào trong bộ lọc này.'),
                        // Mascot only for the genuine "no notifications at
                        // all" case — not for search/filter-empty, to avoid
                        // overusing it on every filter combination.
                        mascotAsset: (_query.isEmpty && all.isEmpty) ? 'mascot_notification_empty.png' : null,
                      )
                    : RefreshIndicator(
                        onRefresh: () async => _load(),
                        child: ListView(
                          padding: const EdgeInsets.fromLTRB(
                            AppSpacing.lg,
                            0,
                            AppSpacing.lg,
                            AppSpacing.lg,
                          ),
                          children: _buildGroupedList(filtered),
                        ),
                      ),
              ),
            ],
          );
        },
      ),
    );
  }

  List<Widget> _buildGroupedList(List<DriverNotification> items) {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final yesterday = today.subtract(const Duration(days: 1));
    final weekAgo = today.subtract(const Duration(days: 7));

    final groups = <String, List<DriverNotification>>{
      'Hôm nay': [],
      'Hôm qua': [],
      '7 ngày qua': [],
      'Cũ hơn': [],
    };

    for (final n in items) {
      final day = DateTime(n.timestamp.year, n.timestamp.month, n.timestamp.day);
      if (day == today) {
        groups['Hôm nay']!.add(n);
      } else if (day == yesterday) {
        groups['Hôm qua']!.add(n);
      } else if (day.isAfter(weekAgo)) {
        groups['7 ngày qua']!.add(n);
      } else {
        groups['Cũ hơn']!.add(n);
      }
    }

    final widgets = <Widget>[];
    groups.forEach((label, entries) {
      if (entries.isEmpty) return;
      widgets.add(_GroupHeader(label: label));
      for (final n in entries) {
        widgets.add(
          Padding(
            padding: const EdgeInsets.only(bottom: AppSpacing.sm),
            child: NotificationTile(notification: n, onTap: () => _markRead(n)),
          ),
        );
      }
    });
    return widgets;
  }
}

class _GroupHeader extends StatelessWidget {
  const _GroupHeader({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: AppSpacing.md, bottom: AppSpacing.sm),
      child: Text(
        label,
        style: Theme.of(context).textTheme.labelMedium?.copyWith(color: AppColors.textSecondary),
      ),
    );
  }
}

class _FilterChip extends StatelessWidget {
  const _FilterChip({required this.label, required this.selected, required this.onTap});

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: AppSpacing.sm),
      child: GestureDetector(
        onTap: onTap,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 180),
          padding: const EdgeInsets.symmetric(horizontal: 14),
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: selected ? AppColors.primary : AppColors.surfaceAlt,
            borderRadius: AppRadius.pillAll,
            border: Border.all(color: selected ? AppColors.primary : AppColors.border),
          ),
          child: Text(
            label,
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: selected ? Colors.white : AppColors.textSecondary,
            ),
          ),
        ),
      ),
    );
  }
}
