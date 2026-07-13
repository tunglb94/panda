import 'package:flutter/material.dart';

/// A voucher's redemption state, driving both its badge label/color and
/// whether it can be selected.
///
/// There is no Voucher backend yet — `backend/services/promotion` has real
/// domain logic (see `PromotionRule`/`VoucherValidator`) but is not wired to
/// any gRPC handler or REST route, so no client can call it. This enum and
/// [Voucher] exist so the UI is fully built and ready; the rider's actual
/// voucher list is empty today (see `VoucherRepository`/`voucher_list_sheet.dart`)
/// because there is nothing real to show — not because this widget can't
/// render one.
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

/// One voucher/promotion campaign as shown to the rider.
///
/// Field shape intentionally mirrors what `backend/services/promotion`'s
/// `Voucher` entity already models (code, discount, min order, expiry,
/// budget) so wiring this to a real API later is a data-mapping change, not
/// a UI redesign.
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

  final String id;
  final String code;
  final String title;
  final String description;
  final IconData icon;
  final Color accentColor;

  /// Short display value, e.g. "-20%" or "-30,000đ".
  final String discountLabel;

  final VoucherStatus status;

  /// e.g. "Đơn tối thiểu 50,000đ · Chỉ áp dụng xe máy".
  final String? conditionText;
  final DateTime? expiresAt;

  /// 0.0-1.0 campaign budget consumed, when the backend exposes it. Null
  /// hides the progress row entirely rather than showing a fake 0%.
  final double? budgetUsedRatio;

  /// Display-only percent discount — there is no Promotion Engine API to
  /// apply this to a real fare (see class doc comment); the backend's
  /// EstimateFare response never carries a discount, so this value is
  /// never used in any fare calculation, only shown in the voucher's own
  /// badge/detail UI.
  final int discountPercent;
}
