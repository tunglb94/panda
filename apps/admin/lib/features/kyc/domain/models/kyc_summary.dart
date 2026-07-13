/// The 4 dashboard cards (Phần 10): Pending/Approved/Rejected/Expired counts
/// of driver verifications.
class KYCSummary {
  const KYCSummary({
    required this.pending,
    required this.approved,
    required this.rejected,
    required this.expired,
  });

  final int pending;
  final int approved;
  final int rejected;
  final int expired;

  factory KYCSummary.fromJson(Map<String, dynamic> json) => KYCSummary(
        pending: json['pending'] as int? ?? 0,
        approved: json['approved'] as int? ?? 0,
        rejected: json['rejected'] as int? ?? 0,
        expired: json['expired'] as int? ?? 0,
      );
}
