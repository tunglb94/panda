import 'package:flutter/material.dart';

/// Colored status pill (Phần 1's Trạng thái column) — pending/under_review
/// are both shown as "amber" (both mean "chờ duyệt" from an admin's
/// perspective; under_review only exists transiently during a review call).
class StatusBadge extends StatelessWidget {
  const StatusBadge({super.key, required this.status});

  final String status;

  static const _labels = {
    'pending': 'Chờ duyệt',
    'under_review': 'Đang duyệt',
    'approved': 'Đã duyệt',
    'rejected': 'Từ chối',
    'expired': 'Hết hạn',
  };

  static const _colors = {
    'pending': Color(0xFFF59E0B),
    'under_review': Color(0xFFF59E0B),
    'approved': Color(0xFF16A34A),
    'rejected': Color(0xFFDC2626),
    'expired': Color(0xFF6B7280),
  };

  @override
  Widget build(BuildContext context) {
    final color = _colors[status] ?? const Color(0xFF6B7280);
    final label = _labels[status] ?? status;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: color.withValues(alpha: 0.4)),
      ),
      child: Text(label, style: TextStyle(color: color, fontSize: 12, fontWeight: FontWeight.w600)),
    );
  }
}
