/// Mirrors the backend's `entity.RiderKYCStatus`
/// (backend/services/user/domain/entity/rider_verification.go).
enum RiderKYCStatus {
  pending,
  approved,
  rejected;

  static RiderKYCStatus fromWire(String value) => switch (value) {
        'approved' => RiderKYCStatus.approved,
        'rejected' => RiderKYCStatus.rejected,
        _ => RiderKYCStatus.pending,
      };

  String get label => switch (this) {
        RiderKYCStatus.pending => 'Đang chờ duyệt',
        RiderKYCStatus.approved => 'Đã duyệt',
        RiderKYCStatus.rejected => 'Bị từ chối',
      };

  bool get isApproved => this == RiderKYCStatus.approved;
  bool get isRejected => this == RiderKYCStatus.rejected;
}

class RiderVerification {
  const RiderVerification({
    required this.fullName,
    required this.nationalIdNumber,
    required this.dateOfBirth,
    required this.status,
    required this.cccdFrontUploaded,
    required this.cccdBackUploaded,
    required this.rejectReason,
  });

  factory RiderVerification.fromJson(Map<String, dynamic> json) {
    final dobRaw = json['date_of_birth'] as String?;
    return RiderVerification(
      fullName: json['full_name'] as String? ?? '',
      nationalIdNumber: json['national_id_number'] as String? ?? '',
      dateOfBirth: dobRaw != null ? DateTime.tryParse(dobRaw) : null,
      status: RiderKYCStatus.fromWire(json['status'] as String? ?? 'pending'),
      cccdFrontUploaded: json['cccd_front_uploaded'] as bool? ?? false,
      cccdBackUploaded: json['cccd_back_uploaded'] as bool? ?? false,
      rejectReason: json['reject_reason'] as String? ?? '',
    );
  }

  final String fullName;
  final String nationalIdNumber;
  final DateTime? dateOfBirth;
  final RiderKYCStatus status;
  final bool cccdFrontUploaded;
  final bool cccdBackUploaded;
  final String rejectReason;
}
