import 'package:flutter/material.dart';

import 'package:rider/shared/utils/currency_format.dart';

/// A voucher's redemption state, driving both its badge label/color and
/// whether it can be selected.
enum VoucherStatus { available, applied, unavailable, used, expired }

extension VoucherStatusPresentation on VoucherStatus {
  String get badgeLabel => switch (this) {
        VoucherStatus.available => 'Có thể dùng',
        VoucherStatus.applied => 'Đang áp dụng',
        VoucherStatus.unavailable => 'Không khả dụng',
        VoucherStatus.used => 'Đã dùng',
        VoucherStatus.expired => 'Hết hạn',
      };

  bool get isSelectable => this == VoucherStatus.available || this == VoucherStatus.applied;
}

/// One voucher/promotion campaign as shown to the rider — backed by the
/// real `GET /api/v1/rider/vouchers` endpoint (gateway's `PromotionHandler`,
/// `backend/services/promotion`'s Voucher entity). [fromApi] is the only
/// production constructor; the base constructor stays available for
/// building a locally-mutated copy (see `voucher_list_sheet.dart`'s
/// "selected" clone).
class Voucher {
  const Voucher({
    required this.id,
    required this.code,
    required this.title,
    required this.description,
    required this.icon,
    required this.accentColor,
    required this.discountLabel,
    required this.status,
    this.conditionText,
    this.expiresAt,
    this.budgetUsedRatio,
    this.discountPercent = 0,
  });

  factory Voucher.fromApi(Map<String, dynamic> json, {required VoucherStatus status}) {
    final discountType = json['discount_type'] as String? ?? 'percentage';
    final discountValue = (json['discount_value'] as num?)?.toInt() ?? 0;
    final maxDiscount = (json['max_discount'] as num?)?.toInt() ?? 0;
    final minOrder = (json['min_order'] as num?)?.toInt() ?? 0;
    final endTimeRaw = json['end_time'] as String?;

    final discountLabel = discountType == 'flat'
        ? '-${formatMoney(discountValue, 'VND')}'
        : '-$discountValue%${maxDiscount > 0 ? ' (tối đa ${formatMoney(maxDiscount, 'VND')})' : ''}';

    return Voucher(
      id: json['id'] as String? ?? '',
      code: json['code'] as String? ?? '',
      title: json['name'] as String? ?? (json['voucher_name'] as String? ?? ''),
      description: json['description'] as String? ?? '',
      icon: Icons.local_offer_outlined,
      accentColor: const Color(0xFF1A8C4E),
      discountLabel: discountLabel,
      status: status,
      conditionText: minOrder > 0 ? 'Đơn tối thiểu ${formatMoney(minOrder, 'VND')}' : null,
      expiresAt: endTimeRaw != null ? DateTime.tryParse(endTimeRaw) : null,
      discountPercent: discountType == 'percentage' ? discountValue : 0,
    );
  }

  final String id;
  final String code;
  final String title;
  final String description;
  final IconData icon;
  final Color accentColor;

  /// Short display value, e.g. "-20%" or "-30,000đ".
  final String discountLabel;

  final VoucherStatus status;

  /// e.g. "Đơn tối thiểu 50,000đ".
  final String? conditionText;
  final DateTime? expiresAt;

  /// 0.0-1.0 campaign budget consumed, when the backend exposes it. Null
  /// hides the progress row entirely rather than showing a fake 0%.
  final double? budgetUsedRatio;

  /// Display-only percent discount, shown in the voucher's own badge/detail
  /// UI — the actual applied discount amount always comes from
  /// `PromotionRepository.apply`, never computed client-side.
  final int discountPercent;
}
