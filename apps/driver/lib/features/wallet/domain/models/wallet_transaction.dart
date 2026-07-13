/// One row of the driver's own wallet ledger (Phần 6/9) — Ride/Delivery
/// income, Commission, Bonus, Withdrawal, Refund, Adjustment, ...
class WalletTransaction {
  const WalletTransaction({
    required this.type,
    required this.direction,
    required this.amountCents,
    required this.currency,
    required this.description,
    required this.paymentMethod,
    required this.createdAt,
  });

  final String type;
  final String direction; // "credit" | "debit"
  final int amountCents;
  final String currency;
  final String description;
  final String paymentMethod; // "cash" | "wallet" | ""
  final DateTime? createdAt;

  bool get isCredit => direction == 'credit';

  factory WalletTransaction.fromJson(Map<String, dynamic> json) => WalletTransaction(
        type: json['type'] as String? ?? '',
        direction: json['direction'] as String? ?? 'credit',
        amountCents: (json['amount_cents'] as num?)?.toInt() ?? 0,
        currency: json['currency'] as String? ?? 'VND',
        description: json['description'] as String? ?? '',
        paymentMethod: json['payment_method'] as String? ?? '',
        createdAt: DateTime.tryParse(json['created_at'] as String? ?? ''),
      );

  Map<String, dynamic> toJson() => {
        'type': type,
        'direction': direction,
        'amount_cents': amountCents,
        'currency': currency,
        'description': description,
        'payment_method': paymentMethod,
        'created_at': createdAt?.toIso8601String() ?? '',
      };
}

/// Vietnamese label for a wallet transaction type — Phần 6's category list.
String walletTransactionTypeLabel(String type) => switch (type) {
      'ride_income' => 'Thu nhập chuyến xe',
      'delivery_income' => 'Thu nhập giao hàng',
      'commission' => 'Hoa hồng',
      'platform_receivable' => 'Panda phải thu',
      'platform_payable' => 'Panda phải trả',
      'promotion_subsidy' => 'Panda bù khuyến mãi',
      'voucher_subsidy' => 'Panda bù voucher',
      'bonus' => 'Thưởng',
      'penalty' => 'Phạt',
      'withdrawal' => 'Rút tiền',
      'refund' => 'Hoàn tiền',
      'adjustment' || 'manual_credit' || 'manual_debit' => 'Điều chỉnh',
      'cash_collected' => 'Thu tiền mặt',
      _ => type,
    };
