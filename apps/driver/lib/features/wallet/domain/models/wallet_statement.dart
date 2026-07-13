/// Phần 9 — Driver Statement for a date range (Ngày/Tuần/Tháng), per-category totals.
class WalletStatement {
  const WalletStatement({
    required this.currency,
    required this.rideIncomeCents,
    required this.deliveryIncomeCents,
    required this.commissionCents,
    required this.promotionCents,
    required this.voucherCents,
    required this.cashIncomeCents,
    required this.electronicIncomeCents,
    required this.withdrawalCents,
    required this.outstandingCents,
  });

  final String currency;
  final int rideIncomeCents;
  final int deliveryIncomeCents;
  final int commissionCents;
  final int promotionCents;
  final int voucherCents;
  final int cashIncomeCents;
  final int electronicIncomeCents;
  final int withdrawalCents;
  final int outstandingCents;

  /// Total driver income for the period (Ride + Delivery) — the figure the
  /// Finance Dashboard's Hôm nay/Tuần/Tháng cards show (Phần 7).
  int get totalIncomeCents => rideIncomeCents + deliveryIncomeCents;

  static const empty = WalletStatement(
    currency: 'VND',
    rideIncomeCents: 0,
    deliveryIncomeCents: 0,
    commissionCents: 0,
    promotionCents: 0,
    voucherCents: 0,
    cashIncomeCents: 0,
    electronicIncomeCents: 0,
    withdrawalCents: 0,
    outstandingCents: 0,
  );

  factory WalletStatement.fromJson(Map<String, dynamic> json) => WalletStatement(
        currency: json['currency'] as String? ?? 'VND',
        rideIncomeCents: (json['ride_income_cents'] as num?)?.toInt() ?? 0,
        deliveryIncomeCents: (json['delivery_income_cents'] as num?)?.toInt() ?? 0,
        commissionCents: (json['commission_cents'] as num?)?.toInt() ?? 0,
        promotionCents: (json['promotion_cents'] as num?)?.toInt() ?? 0,
        voucherCents: (json['voucher_cents'] as num?)?.toInt() ?? 0,
        cashIncomeCents: (json['cash_income_cents'] as num?)?.toInt() ?? 0,
        electronicIncomeCents: (json['electronic_income_cents'] as num?)?.toInt() ?? 0,
        withdrawalCents: (json['withdrawal_cents'] as num?)?.toInt() ?? 0,
        outstandingCents: (json['outstanding_cents'] as num?)?.toInt() ?? 0,
      );
}
