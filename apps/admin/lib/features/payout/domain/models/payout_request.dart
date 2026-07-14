/// Admin-facing Payout Request shape — mirrors gateway's `payoutRequestJSON`
/// (backend/services/gateway/http/handlers/wallet_handler.go, admin fields
/// added in admin_wallet_handler.go's ListPayoutRequests).
class PayoutRequest {
  const PayoutRequest({
    required this.id,
    required this.driverId,
    required this.amount,
    required this.currency,
    required this.bankName,
    required this.maskedAccountNumber,
    required this.status,
    required this.requestedAt,
    this.reviewedAt,
    this.rejectReason,
    this.paidAt,
  });

  factory PayoutRequest.fromJson(Map<String, dynamic> json) => PayoutRequest(
        id: json['payout_request_id'] as String? ?? '',
        driverId: json['driver_id'] as String? ?? '',
        amount: (json['amount_cents'] as num?)?.toInt() ?? 0,
        currency: json['currency'] as String? ?? 'VND',
        bankName: json['bank_name'] as String? ?? '',
        maskedAccountNumber: json['masked_account_number'] as String? ?? '',
        status: json['status'] as String? ?? 'pending',
        requestedAt: json['requested_at'] as String? ?? '',
        reviewedAt: json['reviewed_at'] as String?,
        rejectReason: json['reject_reason'] as String?,
        paidAt: json['paid_at'] as String?,
      );

  final String id;
  final String driverId;
  final int amount; // smallest currency unit — whole VND, same convention as Voucher budget fields
  final String currency;
  final String bankName;
  final String maskedAccountNumber;
  final String status; // pending | approved | rejected | paid
  final String requestedAt; // RFC3339
  final String? reviewedAt;
  final String? rejectReason;
  final String? paidAt;
}
