/// A driver's withdrawal request (Phần 5/8) — Pending -> Approved -> Paid,
/// or Pending -> Rejected. No real money movement — "Chỉ mô phỏng."
class PayoutRequest {
  const PayoutRequest({
    required this.payoutRequestId,
    required this.amountCents,
    required this.currency,
    required this.bankName,
    required this.maskedAccountNumber,
    required this.status,
    required this.requestedAt,
    this.reviewedAt,
    this.rejectReason = '',
    this.paidAt,
  });

  final String payoutRequestId;
  final int amountCents;
  final String currency;
  final String bankName;
  final String maskedAccountNumber;
  final String status; // pending | approved | rejected | paid
  final DateTime? requestedAt;
  final DateTime? reviewedAt;
  final String rejectReason;
  final DateTime? paidAt;

  factory PayoutRequest.fromJson(Map<String, dynamic> json) => PayoutRequest(
        payoutRequestId: json['payout_request_id'] as String? ?? '',
        amountCents: (json['amount_cents'] as num?)?.toInt() ?? 0,
        currency: json['currency'] as String? ?? 'VND',
        bankName: json['bank_name'] as String? ?? '',
        maskedAccountNumber: json['masked_account_number'] as String? ?? '',
        status: json['status'] as String? ?? 'pending',
        requestedAt: DateTime.tryParse(json['requested_at'] as String? ?? ''),
        reviewedAt: DateTime.tryParse(json['reviewed_at'] as String? ?? ''),
        rejectReason: json['reject_reason'] as String? ?? '',
        paidAt: DateTime.tryParse(json['paid_at'] as String? ?? ''),
      );
}

String payoutStatusLabel(String status) => switch (status) {
      'pending' => 'Đang chờ duyệt',
      'approved' => 'Đã duyệt',
      'rejected' => 'Bị từ chối',
      'paid' => 'Đã chuyển tiền',
      _ => status,
    };
