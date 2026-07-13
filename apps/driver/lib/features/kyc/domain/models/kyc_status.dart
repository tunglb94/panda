/// Mirrors the backend's `entity.KYCStatus` (driver/domain/entity/kyc_status.go)
/// exactly — used by both DriverVerification and VehicleVerification.
enum KYCStatus { pending, underReview, approved, rejected, expired }

extension KYCStatusX on KYCStatus {
  static KYCStatus fromWire(String s) => switch (s) {
        'pending' => KYCStatus.pending,
        'under_review' => KYCStatus.underReview,
        'approved' => KYCStatus.approved,
        'rejected' => KYCStatus.rejected,
        'expired' => KYCStatus.expired,
        _ => KYCStatus.pending,
      };

  /// Phần 6 — Status UI copy.
  String get label => switch (this) {
        KYCStatus.pending => 'Đang chờ duyệt',
        KYCStatus.underReview => 'Đang kiểm tra',
        KYCStatus.approved => 'Đã xác minh',
        KYCStatus.rejected => 'Bị từ chối',
        KYCStatus.expired => 'Đã hết hạn',
      };

  bool get isApproved => this == KYCStatus.approved;
  bool get isRejected => this == KYCStatus.rejected;
  bool get isEditable => this == KYCStatus.pending || this == KYCStatus.rejected;
}
